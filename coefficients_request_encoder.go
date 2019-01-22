package ai_stick_comm

// Converts [224x224 B-plane] [224x224 G-plane] [224x224 R-plane] image to HANDLE_IMAGE message for AI stick
func EncodeCoefficientsFileContents(contents []byte) []byte {
	return append(
		append(makeHeader(), makeCoefficients(contents[0x5c:])...), makePadding()...)
}


func encodeSlot(buffer []byte, offset int, code byte, value byte) {
	buffer[offset] = value
	buffer[offset + 1] = code
}

func makeHeader() []byte {
	header := make([]byte, 0x264)	// 612 bytes
	encodeSlot(header,0x10, 1, 0)
	encodeSlot(header,0x14, 1, 0)
	encodeSlot(header,0x18, 1, 0)
	encodeSlot(header,0x1C, 1, 0)

	encodeSlot(header,0x0E0, 0, 0)
	encodeSlot(header,0x0E4, 0, 0)
	encodeSlot(header,0x0E8, 5, 0)
	encodeSlot(header,0x0EC, 5, 2)

	for i := 0xf0; i < 0x248; i += 4 {
		encodeSlot(header, i, 5, 0)
	}

	encodeSlot(header,0x258, 3, 0)
	encodeSlot(header,0x25C, 3, 0)
	encodeSlot(header,0x260, 3, 0)

	return header
}

func makeCoefficients(data []byte) []byte {
	encoded := make([]byte, len(data) * 4)
	for i := 0; i < len(data); i ++ {
		encodeSlot(encoded, i << 2, 3, data[i])
	}
	return encoded
}

func makePadding() []byte {
	padding := make([]byte, 120648+36+0x5c*4)
	// pad to 4KB with 0x00000300
	for i := 0; i < len(padding); i = i + 4 {
		padding[i+1] = 0x03
	}
	return padding
}

