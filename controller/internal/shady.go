package internal

import "bytes"

var (
	controlBytes = []byte{'\x00',
		'\x01',
		'\x02',
		'\x03',
		'\x04',
		'\x05',
		'\x06',
		'\x07',
		'\x08',
		'\x09',
		'\x0B',
		'\x0C',
		'\x0D',
		'\x0E',
		'\x0F',
		'\x10',
		'\x11',
		'\x12',
		'\x13',
		'\x14',
		'\x15',
		'\x16',
		'\x17',
		'\x18',
		'\x19',
		'\x1A',
		'\x1B',
		'\x1C',
		'\x1D',
		'\x1E',
		'\x1F',
		'\x7F'}
)

// Removes all control bytes from the byte array obtained from docker logs
func ReplaceControlBytes(outBytes []byte) []byte {
	for _, b := range controlBytes {
		outBytes = bytes.ReplaceAll(outBytes, []byte{b}, []byte{})
	}

	return outBytes
}
