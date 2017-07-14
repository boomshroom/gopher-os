package acpi

import (
	"gopheros/kernel"
	"gopheros/kernel/kfmt"
	"gopheros/kernel/mem/pmm"
	"gopheros/kernel/mem/vmm"
	"unsafe"
)

const (
	// RDSP must be located in the physical memory region 0xe0000 to 0xfffff
	rsdpLocationLow uintptr = 0xe0000
	rsdpLocationHi  uintptr = 0xfffff
)

// rsdpDescriptor defines the root system descriptor pointer for ACPI 1.0. This
// is used as the entry-point for parsing ACPI data.
type rsdpDescriptor struct {
	// The signature must contain "RSD PTR " (last byte is a space).
	signature [8]byte

	// A value that when added to the sum of all other bytes in the 32-bit
	// RSDT should result in the value 0.
	checksum uint8

	oemID [6]byte

	// ACPI revision number. It is 0 for ACPI1.0 and 2 for versions 2.0 to 6.1.
	revision uint8

	// Physical address of 32-bit root system descriptor table.
	rsdtAddr uint32
}

// valid calculates the checksum for the rsdp.
func (h *rsdpDescriptor) valid() bool {
	var (
		sum              uint8
		curPtr           = uintptr(unsafe.Pointer(&h.signature[0]))
		sizeofDescriptor = unsafe.Sizeof(*h)
	)

	for i := uintptr(0); i < sizeofDescriptor; i++ {
		sum += *(*uint8)(unsafe.Pointer(curPtr + i))
	}

	return sum == 0
}

// rsdpDescriptor2 extends rsdpDescriptor with additional fields. It is used
// when rsdpDescriptor.revision > 1.
type rsdpDescriptor2 struct {
	rsdpDescriptor

	// The size of the 64-bit root system descriptor table.
	length uint32

	// Physical address of 64-bit root system descriptor table.
	xsdtAddr uint64

	// A value that when added to the sum of all bytes in the 64-bit RSDT
	// should result in the value 0.
	extendedChecksum uint8

	reserved [3]byte
}

func findRSDP() *kernel.Error {
	var (
		signature = [8]byte{'R', 'S', 'D', ' ', 'P', 'T', 'R', ' '}
		rsdp      *rsdpDescriptor
	)

	// Cleanup temporary identity mappings when the function returns
	defer func() {
		for curPage := vmm.PageFromAddress(rsdpLocationLow); curPage <= vmm.PageFromAddress(rsdpLocationHi); curPage++ {
			vmm.Unmap(curPage)
		}
	}()

	// Setup temporary identity mapping so we can scan for the header
	for curPage := vmm.PageFromAddress(rsdpLocationLow); curPage <= vmm.PageFromAddress(rsdpLocationHi); curPage++ {
		if err := vmm.Map(curPage, pmm.Frame(curPage), vmm.FlagPresent|vmm.FlagRW); err != nil {
			return err
		}
	}

	// The RSDP should be aligned on a 16-byte boundary
checkNextBlock:
	for curPtr := rsdpLocationLow; curPtr < rsdpLocationHi; curPtr += 16 {
		rsdp = (*rsdpDescriptor)(unsafe.Pointer(curPtr))
		for i := 0; i < 8; i++ {
			if rsdp.signature[i] != signature[i] {
				continue checkNextBlock
			}
		}
		kfmt.Printf("Found RSDP at 0x%x; ACPI rev: %d\n", curPtr, rsdp.revision)
		kfmt.Printf("OEM ID: %s\n", string(rsdp.oemID[:]))
		kfmt.Printf("Valid? %t\n", rsdp.valid())

	}

	return nil
}
