package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func MergeDumps(target, dir string) (*os.File, error) {
	// todo test

	// contents of data directory => measurements
	log.Println("trying to merge measurements")
	data, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	log.Printf("found %d potential dumps\n", len(data))

	// target file
	t, err := os.Create(target)
	if err != nil {
		return nil, err
	}
	log.Println("created", t.Name())
	write := func(line string) error {
		_, err := t.Write([]byte(fmt.Sprintf("%s\n", line)))
		return err
	}
	err = write("tProducer, tConsumer")
	if err != nil {
		return nil, err
	}

	// go over all dumps and append measurements to target file
	for _, x := range data {
		file, err := os.Open(x.Name())
		if err != nil {
			return nil, err
		}
		defer file.Close() // todo
		log.Println("appending measurements from", file.Name())

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "tProducer, tConsumer" {
				continue
			}
			err = write(line)
			if err != nil {
				return nil, err
			}
		}
	}

	// merge successful
	log.Println("merge successful")
	return t, nil
}
