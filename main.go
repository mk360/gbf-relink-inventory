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
	results := memory.ScanMemory(handle, needle_pattern)

	if len(results) == 0 {
		log.Println("Byte pattern not found in heap memory")
		os.Exit(1)
	} else {
		log.Printf("Found %d matches in memory.\n", len(results))
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(int_desired_quantity))

		itemHexBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(itemHexBuf, item_hex)
		combined := append(itemHexBuf, buf...)
		for _, address := range results {
			memory.WriteMemory(handle, address, combined)
		}
		os.Exit(0)
	}
}
