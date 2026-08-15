package main

import (
	"encoding/binary"
	"fmt"
	"gbf-relink/data"
	"gbf-relink/memory"
	"log"
	"os"
	"strconv"
	"strings"

	windows "golang.org/x/sys/windows"
)

func little_endian_cur_qty(value string) string {
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		log.Fatalln("StringToLittleEndianPatternBytes: invalid uint32 string:", err)
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(parsed))

	tokens := make([]string, len(buf))
	for i, b := range buf {
		tokens[i] = fmt.Sprintf("%02X", b)
	}

	return strings.Join(tokens, " ")
}

func get_byte_array(hex_number uint32) string {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, hex_number)

	tokens := make([]string, len(buf))
	for i, b := range buf {
		tokens[i] = fmt.Sprintf("%02X", b)
	}

	return strings.Join(tokens, " ")
}

func main() {
	pid := memory.GetGameProcessPID()
	handle := memory.GetProcessHandle(pid)
	defer windows.CloseHandle(handle)
	if len(os.Args) < 4 {
		log.Fatalln("Please specify: item, current quantity, desired quantity")
	}
	item_name := os.Args[1]
	item_hex, _ := data.GetItem(item_name)
	if item_hex == 0 {
		log.Fatalln("Item not found in dict")
	}
	item_current_quantity := os.Args[2]
	if _, err := strconv.Atoi(item_current_quantity); err != nil {
		log.Fatalln("Invalid current quantity argument")
	}
	item_desired_quantity := os.Args[3]
	int_desired_quantity, err := strconv.Atoi(item_desired_quantity)
	if err != nil || int_desired_quantity < 0 {
		log.Fatalln("Invalid desired quantity argument")
	}
	needle_pattern := get_byte_array(item_hex) + " " + little_endian_cur_qty(item_current_quantity)
	address := data.FindPattern(pid, handle, needle_pattern)

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(int_desired_quantity))

	itemHexBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(itemHexBuf, item_hex)
	combined := append(itemHexBuf, buf...)
	memory.WriteMemory(handle, address, combined)
	os.Exit(0)
	// address := uintptr(0x00401000) // Target Memory Address
	// size := uint32(4)              // Bytes to read

	// // 1. Open Process

	// // 2. Prepare Buffer
	// buffer := make([]byte, size)

	// // 3. Call ReadProcessMemory
	// kernel32 := syscall.MustLoadDLL("kernel32.dll")
	// procRead := kernel32.MustFindProc("ReadProcessMemory")

	// var bytesRead uint32
	// ret, _, _ := procRead.Call(
	// 	uintptr(handle),
	// 	address,
	// 	uintptr(unsafe.Pointer(&buffer[0])),
	// 	uintptr(size),
	// 	uintptr(unsafe.Pointer(&bytesRead)),
	// )

	// if ret == 0 {
	// 	panic("Failed to read memory")
	// }

	// // 4. Interpret Data
	// // Example: Convert first 4 bytes to uint32 (Little Endian)
	// val := uint32(buffer[0]) | uint32(buffer[1])<<8 | uint32(buffer[2])<<16 | uint32(buffer[3])<<24
	// fmt.Printf("Read value: %d\n", val)
}
