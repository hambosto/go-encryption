package worker

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

func (ws *WorkerStream) writeResults(
	writer io.Writer,
	results <-chan result,
	wg *sync.WaitGroup,
	errChan chan<- error,
) {
	defer wg.Done()

	pending := make(map[uint32]result)
	var nextIndex uint32

	for res := range results {
		if res.err != nil {
			errChan <- fmt.Errorf("processing chunk %d: %w", res.index, res.err)
			return
		}

		pending[res.index] = res

		// Process chunks in order
		for {
			current, exists := pending[nextIndex]
			if !exists {
				break
			}

			if err := ws.writeChunk(writer, current); err != nil {
				errChan <- fmt.Errorf("writing chunk %d: %w", nextIndex, err)
				return
			}

			delete(pending, nextIndex)
			nextIndex++
		}
	}
}

func (ws *WorkerStream) writeChunk(writer io.Writer, res result) error {
	// For encryption, we need to prefix each chunk with its size
	if ws.processor.IsEncryption {
		if err := ws.writeChunkSize(writer, len(res.data)); err != nil {
			return err
		}
	}

	// Write the actual data
	if _, err := writer.Write(res.data); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	// Update the progress bar
	if err := ws.progress.Add(res.size); err != nil {
		return fmt.Errorf("progress update failed: %w", err)
	}

	return nil
}

func (ws *WorkerStream) writeChunkSize(writer io.Writer, size int) error {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(size))

	if _, err := writer.Write(buf[:]); err != nil {
		return fmt.Errorf("chunk size write failed: %w", err)
	}

	return nil
}

func (ws *WorkerStream) readEncryptChunks(reader io.Reader, jobs chan<- job) error {
	var index uint32
	buffer := make([]byte, chunkSize)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read failed: %w", err)
		}

		// Make a copy of the data to avoid buffer reuse issues
		data := make([]byte, n)
		copy(data, buffer[:n])

		// Send job to workers
		jobs <- job{data: data, index: atomic.LoadUint32(&index)}
		atomic.AddUint32(&index, 1)
	}

	return nil
}

func (ws *WorkerStream) readDecryptChunks(reader io.Reader, jobs chan<- job) error {
	var index uint32
	var sizeBuf [4]byte

	for {
		// Read chunk size
		_, err := io.ReadFull(reader, sizeBuf[:])
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("chunk size read failed: %w", err)
		}

		// Decode chunk size
		chunkLen := binary.BigEndian.Uint32(sizeBuf[:])

		// Read chunk data
		data := make([]byte, chunkLen)
		if _, err := io.ReadFull(reader, data); err != nil {
			return fmt.Errorf("chunk data read failed: %w", err)
		}

		// Send job to workers
		jobs <- job{data: data, index: atomic.LoadUint32(&index)}
		atomic.AddUint32(&index, 1)
	}

	return nil
}
