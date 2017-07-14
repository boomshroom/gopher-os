package kmain

import (
	"gopheros/device/acpi"
	"gopheros/kernel"
	"gopheros/kernel/goruntime"
	"gopheros/kernel/hal"
	"gopheros/kernel/hal/multiboot"
	"gopheros/kernel/kfmt"
	"gopheros/kernel/mem/pmm/allocator"
	"gopheros/kernel/mem/vmm"
)

var (
	errKmainReturned = &kernel.Error{Module: "kmain", Message: "Kmain returned"}
)

// Kmain is the only Go symbol that is visible (exported) from the rt0 initialization
// code. This function is invoked by the rt0 assembly code after setting up the GDT
// and setting up a a minimal g0 struct that allows Go code using the 4K stack
// allocated by the assembly code.
//
// The rt0 code passes the address of the multiboot info payload provided by the
// bootloader as well as the physical addresses for the kernel start/end. In
// addition, the start of the kernel virtual address space is passed to the
// kernelPageOffset argument.
//
// Kmain is not expected to return. If it does, the rt0 code will halt the CPU.
//
//go:noinline
func Kmain(multibootInfoPtr, kernelStart, kernelEnd, kernelPageOffset uintptr) {
	multiboot.SetInfoPtr(multibootInfoPtr)

	var err *kernel.Error
	if err = allocator.Init(kernelStart, kernelEnd); err != nil {
		panic(err)
	} else if err = vmm.Init(kernelPageOffset); err != nil {
		panic(err)
	} else if err = goruntime.Init(); err != nil {
		panic(err)
	}

	// After goruntime.Init returns we can safely use defer
	defer func() {
		// Use kfmt.Panic instead of panic to prevent the compiler from
		// treating kernel.Panic as dead-code and eliminating it.
		kfmt.Panic(errKmainReturned)
	}()

	// Detect and initialize hardware
	hal.DetectHardware()

	acpi.Init()
}
