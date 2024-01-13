package buffer

import (
	"benchmark/lib/config"
	"benchmark/lib/utils"
	"errors"
	"fmt"
	"log"
	"os"
)

//
// this buffer is deprecated and using it is discouraged
//

type Measurement struct {
	TProducer int64
	TConsumer int64
}

type Buffer struct {
	Buffer []Measurement
	Size   uint
	ptr    uint
	nFlush uint // number of times this buffer has been flushed (keeping track for consistent file names)
}

// NewBuffer creates a new buffer that can hold up to <size> measurements
func NewBuffer(size uint) *Buffer {
	return &Buffer{
		Buffer: make([]Measurement, size),
		Size:   size,
		ptr:    0,
		nFlush: 0,
	}
}

// Store a single measurement in the buffer
// it checks automatically whether the buffer is full and stores measurements in a file accordingly
func (b *Buffer) Store(workerId string, config *config.Config, measurement Measurement) error {
	// throw error if buffer wasn't set up correctly
	if b.Size == 0 {
		return errors.New("buffer has size 0, cannot store anything")
	}
	// buffer is full
	if b.ptr >= b.Size {
		log.Println(workerId+":", "buffer full, flushing buffer")
		// flush buffer and start with empty buffer & ptr = 0
		err := b.Flush(workerId, config.Experiment.DataDir, fmt.Sprintf("%s-%s-dump-%d.csv", config.Experiment.Id, workerId, b.nFlush))
		if err != nil {
			return err
		}
	}
	// store the new measurement
	b.Buffer[b.ptr] = measurement
	b.ptr++

	return nil
}

// Flush stores the buffer contents in csv format to a file and empties the buffer
// - workerId: just for logging
// - filename: should end in .csv and shouldn't exist already
// - dir: should exist!
func (b *Buffer) Flush(workerId, dir, filename string) error {

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
		if measurement.TProducer == 0 && measurement.TConsumer == 0 {
			log.Println(workerId+":", "buffer not full, flushing early")
			break // break here because the "0, 0" line doesn't need to be written
		}
		n, err = file.Write([]byte(fmt.Sprintf("%d, %d\n", measurement.TProducer, measurement.TConsumer)))
		if err != nil {
			return err
		}
		bytesWritten += n
	}
	log.Printf("%s: flushed %d bytes to %s/%s\n", workerId, bytesWritten, dir, filename)

	// empty buffer, increase flush counter, reset current index in buffer to 0
	b.Buffer = make([]Measurement, b.Size)
	b.nFlush++
	b.ptr = 0

	return nil
}

func (b *Buffer) Close(workerId string, config *config.Config) error {
	var err error
	if b.ptr != 0 {
		err = b.Flush(
			workerId,
			config.Experiment.DataDir,
			fmt.Sprintf("%s-%s-flush-%d.csv", config.Experiment.Id, workerId, b.nFlush),
		)
	}
	return err
}
