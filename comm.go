package ai_stick_comm

import (
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
	"time"
	"unsafe"
)

const (
	SG_IO                = 0x2285
	SG_DXFER_TO_DEV      = -2
	SG_DXFER_FROM_DEV    = -3

	SENSE_BUF_LEN        = 64
	TIMEOUT_10_SECS      = 10000

	READ10               = 0x28
	WRITE10              = 0x2A
)

type SgIoHdr struct {
	InterfaceID    int32
	DxferDirection int32
	CmdLen         uint8
	MxSbLen        uint8
	IovecCount     uint16
	DxferLen       uint32
	Dxferp         *byte
	Cmdp           *uint8
	Sbp            *byte
	Timeout        uint32
	Flags          uint32
	PackID         int32
	pad0           [4]byte
	UsrPtr         *byte
	Status         uint8
	MaskedStatus   uint8
	MsgStatus      uint8
	SbLenWr        uint8
	HostStatus     uint16
	DriverStatus   uint16
	Resid          int32
	Duration       uint32
	Info           uint32
}

func OpenSgDevice(fname string) (*os.File, error) {
	f, err := os.OpenFile(fname, os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func CloseSgDevice(device *os.File) {
	_ = device.Close()
}

func MakeOutputBuffer() []byte {
	return make([]byte, 35389440)
}


func ReadOrWrite10(f *os.File, operation uint8, direction int32, buffer []uint8, offset uint32, size uint32) error {
	senseBuf := make([]byte, SENSE_BUF_LEN)

	// READ10 command:
	// 00		operation code (READ10=0x28, WRITE10=0x2A)
	// 01   	various bits
	// 02..05   logical block address (MSB -> LSB)
	// 06		group number
	// 07..08   transfer length (MSB -> LSB)
	// 09		control

	blocks := size >> 9
	command := []uint8{
		operation,
		0,
		0, 0,  0x20, 0,
		0,
		uint8(blocks >> 8), uint8(blocks & 0xFF),
		0,
	}

	ioHdr := &SgIoHdr{
		InterfaceID:    int32('S'),
		CmdLen:         uint8(len(command)),
		MxSbLen:        SENSE_BUF_LEN,
		DxferDirection: direction,
		Cmdp:           &command[0],
		Sbp:            &senseBuf[0],
		Timeout:        TIMEOUT_10_SECS,
		Duration:       12,
		Dxferp:			&buffer[offset],
		DxferLen:		size,
	}

	err := SgioSyscall(f, ioHdr)
	if err != nil {
		return err
	}

	return nil
}


func ReadOrWriteBigBuffer(f *os.File, operation uint8, direction int32, buffer []uint8) error {
	length := uint32(len(buffer))
	var offset uint32 = 0
	for {
		size := length - offset
		if size > 1048576 {
			size = 1048576
		}

		err := ReadOrWrite10(f, operation, direction, buffer, offset, size)
		if err != nil {
			return err
		}

		offset += size
		if offset >= length {
			return nil
		}
	}
}


func Write(device *os.File, bytes []byte) error {
	return ReadOrWriteBigBuffer(device, WRITE10, SG_DXFER_TO_DEV, bytes)
}

func Read(device *os.File, output []byte) error {
	return ReadOrWriteBigBuffer(device, READ10, SG_DXFER_FROM_DEV, output)
}


func SgioSyscall(f *os.File, i *SgIoHdr) error {
	return ioctl(f.Fd(), SG_IO, uintptr(unsafe.Pointer(i)))
}

func ioctl(fd, cmd, ptr uintptr) error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, ptr)
	if err != 0 {
		return err
	}
	return nil
}


func main() {
	device, _ := OpenSgDevice("/dev/sg2")

	coefficients, _ := ioutil.ReadFile("_coefficients.dat")
	_ = Write(device, coefficients)

	image, _ := ioutil.ReadFile("_image.dat")
	_ = Write(device, image)

	time.Sleep(1000 * time.Millisecond)

	output := MakeOutputBuffer()

	_ = Read(device, output)

	e4 := ioutil.WriteFile("output.dat", output, 0644)
	fmt.Println("Result", e4)

	CloseSgDevice(device)
}

