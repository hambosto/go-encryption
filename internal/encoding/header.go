package encoding

import (
	"encoding/binary"
)

func addHeader(data []byte) []byte {
	dataWithHeader := make([]byte, headerLength+len(data))
	binary.BigEndian.PutUint32(dataWithHeader, uint32(len(data)))
	copy(dataWithHeader[headerLength:], data)
	return dataWithHeader
}
