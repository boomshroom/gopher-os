package acpi

import (
	"gopheros/kernel"
	"unsafe"
)

// sdtHeader defines the common header for all ACPI-related tables.
type sdtHeader struct {
	// The signature defines the table type.
	signature [4]byte

	// The length of the table
	length uint32

	revision uint8

	// A value that when added to the sum of all other bytes in the table
	// should result in the value 0.
	checksum uint8

	oemID       [6]byte
	oemTableID  [8]byte
	oemRevision uint32

	creatorID       uint32
	creatorRevision uint32
}

// valid calculates the checksum for the table and its contents and returns
// true if the table is valid.
func (h *sdtHeader) valid() bool {
	var (
		sum    uint8
		curPtr = uintptr(unsafe.Pointer(&h.signature[0]))
	)

	for i := uint32(0); i < h.length; i++ {
		sum += *(*uint8)(unsafe.Pointer(curPtr + uintptr(i)))
	}

	return sum == 0
}

func Init() *kernel.Error {
	return findRSDP()
}
