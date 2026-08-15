package data

import (
	"encoding/binary"
	"gbf-relink/memory"
	"log"

	"golang.org/x/sys/windows"
)

var item_list = []uint32{}

func build_item_list(handle windows.Handle, starting_address uintptr) int {
	var slot = 0
	for {
		newAddress := starting_address + uintptr(slot*96)
		var itemID_bytes = memory.ReadMemory(handle, uintptr(newAddress), 4)
		var itemID = binary.BigEndian.Uint32(itemID_bytes)
		_, ok := ItemsByHex[itemID]
		if ok {
			item_list = append(item_list, itemID)
		} else {
			log.Printf("%x\n", newAddress)
			break
		}
		slot++
	}
	return len(item_list)
}
