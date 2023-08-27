package wal

func DeserializeRecord(recordData []byte) (int64, []byte) {
	// The first bytes correspond to the variable-length encoded block ID
	blockID, blockIDLen := decodeVarInt64(recordData)
	data := recordData[blockIDLen:]
	return blockID, data
}

func decodeVarInt64(encoded []byte) (int64, int) {
	var value int64
	var shift uint

	for i, b := range encoded {
		value |= int64(b&0x7F) << shift
		shift += 7
		if b&0x80 == 0 {
			return value, i + 1
		}
	}

	return value, len(encoded)
}
