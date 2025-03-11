package padding

import (
	"encoding/binary"
)

func Pad(data []byte) []byte {
	sizeHeader := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeHeader, uint32(len(data)))

	alignedSize := (len(data) + 15) & ^15
	paddedData := make([]byte, alignedSize)
	copy(paddedData, data)

	return append(sizeHeader, paddedData...)
}
