package encoding

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/vivint/infectious"
)

func prepareDataForEncoding(data []byte) ([]byte, error) {
	if err := validateDataSize(data); err != nil {
		return nil, err
	}

	sizePrefix := make([]byte, 4)
	binary.BigEndian.PutUint32(sizePrefix, uint32(len(data)))
	return append(sizePrefix, data...), nil
}

func encodeWithFEC(fec *infectious.FEC, paddedData []byte) ([]byte, error) {
	buffer := &bytes.Buffer{}
	if err := fec.Encode(paddedData, func(s infectious.Share) { buffer.Write(s.Data) }); err != nil {
		return nil, fmt.Errorf("failed to encode data: %w", err)
	}
	return buffer.Bytes(), nil
}
