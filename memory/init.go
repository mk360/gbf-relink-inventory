package memory

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var PROCESS_NAME = "granblue_fantasy_relink.exe"

type ProcessInfo struct {
	PID  uint32
	Name string
}

func listProcesses() ([]ProcessInfo, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, fmt.Errorf("CreateToolhelp32Snapshot: %w", err)
	}
	defer windows.CloseHandle(snapshot)
	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	if err := windows.Process32First(snapshot, &entry); err != nil {
		return nil, fmt.Errorf("Process32First: %w", err)
	}
	var processes []ProcessInfo
	for {
		processes = append(processes, ProcessInfo{
			PID:  entry.ProcessID,
			Name: windows.UTF16ToString(entry.ExeFile[:]),
		})

		if err := windows.Process32Next(snapshot, &entry); err != nil {
			if err == windows.ERROR_NO_MORE_FILES {
				break
			}
			return nil, fmt.Errorf("Process32Next: %w", err)
		}
	}
	return processes, nil
}

func getProcessByName(name string) (uint32, error) {
	processList, _ := listProcesses()
	for _, process := range processList {
		if process.Name == name {
			return process.PID, nil
		}
	}
	return 0, fmt.Errorf("Process not found")
}

func GetGameProcessPID() uint32 {
	pid, err := getProcessByName(PROCESS_NAME)
	if err != nil {
		log.Fatalln(err)
	}
	return pid
}

func GetProcessHandle(pid uint32) windows.Handle {
	handle, err := windows.OpenProcess(
		windows.PROCESS_VM_READ|windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_WRITE,
		false,
		uint32(pid),
	)
	if err != nil {
		panic(err)
	}
	return handle
}

type ModuleInfo struct {
	BaseAddress uintptr
	Size        uintptr
}

const memMEM_PRIVATE = 0x20000

func searchForPattern(handle windows.Handle, module ModuleInfo, pattern []patternByte, final_address_channel chan uintptr) {
	buf := ReadMemory(handle, module.BaseAddress, module.Size)
	bufLen := len(buf)
	for i := 0; i <= bufLen-len(pattern); i++ {
		match := true
		for j := range len(pattern) {
			if pattern[j].wildcard {
				continue
			}
			if buf[i+j] != pattern[j].value {
				match = false
				break
			}
		}
		if match {
			select {
			case final_address_channel <- module.BaseAddress + uintptr(i):
			default:
			}
		}
	}
}

func ScanMemory(handle windows.Handle, pattern string) (uintptr, error) {
	log.Println("Searching for pattern", pattern)
	pat := parsePattern(pattern)
	patLen := len(pat)
	final_address_channel := make(chan uintptr, 1)
	job_channel := make(chan ModuleInfo)
	wg := sync.WaitGroup{}
	address := uintptr(0)
	maxAddress := uintptr(0x7FFFFFFEFFFF)
	var modulesList = []ModuleInfo{}

	for address < maxAddress {
		var mbi windows.MemoryBasicInformation

		if err := windows.VirtualQueryEx(handle, address, &mbi, 48); err != nil {
			// can't query further
			// stop walking rather than looping forever
			break
		}

		isHeap := mbi.State == windows.MEM_COMMIT &&
			mbi.Type == memMEM_PRIVATE &&
			mbi.Protect == windows.PAGE_READWRITE

		if isHeap && mbi.RegionSize >= uintptr(patLen) {
			modulesList = append(modulesList, ModuleInfo{
				BaseAddress: mbi.BaseAddress,
				Size:        mbi.RegionSize,
			})
		}
		address = mbi.BaseAddress + mbi.RegionSize
	}

	for range runtime.NumCPU() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for module := range job_channel {
				searchForPattern(handle, module, pat, final_address_channel)
			}
		}()
	}

	for _, module := range modulesList {
		job_channel <- module
	}
	close(job_channel)
	wg.Wait()
	select {
	case addr := <-final_address_channel:
		return addr, nil
	default:
		return 0, fmt.Errorf("pattern not found in any heap region")
	}
}
