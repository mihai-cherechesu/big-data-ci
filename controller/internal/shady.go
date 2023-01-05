package internal

import "bytes"

func ReplaceControlBytes(outBytes []byte) []byte {
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x00'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x01'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x02'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x03'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x04'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x05'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x06'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x07'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x08'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x09'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x0B'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x0C'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x0D'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x0E'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x0F'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x10'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x11'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x12'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x13'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x14'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x15'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x16'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x17'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x18'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x19'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x1A'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x1B'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x1C'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x1D'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x1E'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x1F'}, []byte{})
	outBytes = bytes.ReplaceAll(outBytes, []byte{'\x7F'}, []byte{})

	return outBytes
}
