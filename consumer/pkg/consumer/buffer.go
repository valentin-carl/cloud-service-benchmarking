package consumer

import (
	"consumer/pkg/config"
	"consumer/pkg/utils"
	"errors"
	"fmt"
	"log"
	"os"
)

type Buffer struct {
	Buffer []struct {
		tProducer int64
		tConsumer int64
	}
	Size   uint
	ptr    uint
	nDumps uint // number of dumps this buffer has done (keeping track for consistent file names)
}

// NewBuffer creates a new buffer that can hold up to <size> measurements
func NewBuffer(size uint) *Buffer {
	return &Buffer{
		Buffer: make([]struct {
			tProducer int64
			tConsumer int64
		}, size),
		Size:   size,
		ptr:    0,
		nDumps: 0,
	}
}

// Store a single measurement in the buffer
// it checks automatically whether the buffer is full and stores measurements in a file accordingly
func (b *Buffer) Store(workerId string, config *config.Config, measurement struct {
	tProducer int64
	tConsumer int64
}) error {
	// throw error if buffer wasn't set up correctly
	if b.Size == 0 {
		return errors.New("buffer has size 0, cannot store anything")
	}
	// buffer is full
	if b.ptr >= b.Size {
		log.Println(workerId+":", "buffer full, dumping measurements")
		// dump buffer contents and start with empty buffer & ptr = 0
		err := b.Dump(workerId, config.Experiment.DataDir, fmt.Sprintf("%s-%s-%d", config.Experiment.Id, workerId, b.nDumps))
		if err != nil {
			return err
		}
	}
	// store the new measurement
	b.Buffer[b.ptr] = measurement
	b.ptr++

	return nil
}

// Dump stores the buffer contents in csv format to a file and empties the buffer
// - workerId: just for logging
// - filename: should end in .csv and shouldn't exist already
// - dir: should exist!
func (b *Buffer) Dump(workerId, dir, filename string) error {

	// todo validate inputs: check dir exists and filename doesn't

	// create file
	file, err := os.Create(fmt.Sprintf("%s/%s", dir, filename))
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		utils.Handle(file.Close())
	}(file)

	// store buffer contents
	// write column titles
	bytesWritten := 0
	n, err := file.Write([]byte("tProducer, tConsumer\n"))
	if err != nil {
		return err
	}
	bytesWritten += n
	// write measurements
	for _, measurement := range b.Buffer {
		// check whether buffer is not completely full
		if measurement.tProducer == 0 && measurement.tConsumer == 0 {
			log.Println(workerId+":", "buffer not full, dumping early")
			break // break here because the "0, 0" line doesn't need to be written
		}
		n, err = file.Write([]byte(fmt.Sprintf("%d, %d\n", measurement.tProducer, measurement.tConsumer)))
		if err != nil {
			return err
		}
		bytesWritten += n
	}
	log.Printf("%s: dumped %d bytes to %s/%s\n", workerId, bytesWritten, dir, filename)

	// empty buffer, increase dump counter, reset current index in buffer to 0
	b.Buffer = make([]struct {
		tProducer int64
		tConsumer int64
	}, b.Size)
	b.nDumps++
	b.ptr = 0

	return nil
}
