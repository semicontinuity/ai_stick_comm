package ai_stick_comm

func encodeSlot(buffer []byte, offset int, code byte, value byte) {
	buffer[offset] = value
	buffer[offset + 1] = code
}
