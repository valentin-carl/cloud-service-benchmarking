package buffer

import (
	"consumer/pkg/utils"
	"log"
	"os/exec"
	"runtime"
	"strconv"
)

const (
	defaultBlockSize = 4096
	measurementSize  = 128 // 2x int64
)

// CalcOptimalBufferSize uses the systems block size to calculate
// how many elements a buffer can hold to fill exactly one block
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
		return -1, err
	}
	bs, err := strconv.Atoi(string(output))
	log.Printf("block size %d\n", bs)
	return uint(bs), err
}
