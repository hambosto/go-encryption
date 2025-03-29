package worker

// ChunkSize defines the size of data chunks to process
const ChunkSize = 1024 * 1024 // 1MB

// Job represents a chunk of data to be processed
type Job struct {
	Data  []byte
	Index uint32
}

// Result represents the outcome of processing a job
type Result struct {
	Index uint32
	Data  []byte
	Size  int
	Error error
}
