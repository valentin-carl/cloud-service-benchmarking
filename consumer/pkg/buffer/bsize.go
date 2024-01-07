package buffer

import (
	"benchmark/lib/utils"
	"log"
	"os/exec"
	"runtime"
)

const (
	defaultBlockSize = 4096  // bytes
	measurementSize  = 2 * 8 // 2x int64 == 16 bytes
)

// CalcOptimalBufferSize uses the system's block size to calculate
// how many elements a buffer can hold to fill exactly one block
// TODO subtract csv column titles
func CalcOptimalBufferSize() uint {
	bs, err := GetBlockSize()
	utils.Handle(err)
	return bs / measurementSize
}

// GetBlockSize tries to find the system's block size
// linux: stat -fc %s config.json
// macos: stat -f %k config.json
func GetBlockSize() (uint, error) {
	// get correct command depending on os
	var cmd *exec.Cmd
	if os := runtime.GOOS; os == "darwin" {
		log.Println("macos detected")
		cmd = exec.Command("stat", "-f", "%k", "config.json")
	} else if os == "linux" {
		log.Println("linux detected")
		cmd = exec.Command("stat", "-fc", "%s", "config.json")
	} else {
		log.Println("didn't detect valid OS, using default block size")
		return defaultBlockSize, nil
	}

	// execute the command
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	bs, err := utils.ExtractNumber(output)
	log.Printf("block size %d\n", bs) // "block size 0" is usually an error
	return uint(bs), err
}
