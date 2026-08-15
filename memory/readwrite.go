package memory

import (
	"log"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var kernel32 = syscall.MustLoadDLL("kernel32.dll")

func ReadMemory(handle windows.Handle, address uintptr, size uintptr) []byte {
	buffer := make([]byte, size)
	procRead := kernel32.MustFindProc("ReadProcessMemory")
	var bytesRead uintptr = 0
	ret, _, err := procRead.Call(
		uintptr(handle),
		address,
		uintptr(unsafe.Pointer(&buffer[0])),
		size,
		uintptr(unsafe.Pointer(&bytesRead)),
	)
	if ret == 0 {
		log.Println("ReadProcessMemory failed:", err)
	}
	return buffer
}

func WriteMemory(handle windows.Handle, address uintptr, data []byte) {
	dataLen := len(data)
	procWrite := kernel32.MustFindProc("WriteProcessMemory")
	_, _, err := procWrite.Call(
		uintptr(handle),
		address,
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(dataLen),
		uintptr(unsafe.Pointer(&dataLen)),
	)
	log.Println(err)
}
