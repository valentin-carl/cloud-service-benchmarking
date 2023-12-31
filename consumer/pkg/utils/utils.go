package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
)

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}

// MergeMeasurements takes all measurement files and joins them to a single .csv file
// careful: if a file with the same name already exists, it is overwritten and data might be lost
func MergeMeasurements(target, dataDir, outDir string) (*os.File, error) {

	// contents of data directory => measurements
	log.Println("trying to merge measurements")
	data, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}
	log.Printf("found %d potential measurement files\n", len(data))

	// TODO filter data

	// target file
	t, err := os.Create(fmt.Sprintf("%s/%s", outDir, target))
	if err != nil {
		return nil, err
	}
	log.Println("created", t.Name())
	write := func(line string) error {
		_, err := t.Write([]byte(fmt.Sprintf("%s\n", line)))
		log.Println("writing")
		return err
	}
	err = write("tProducer, tConsumer")
	if err != nil {
		log.Println("error")
		return nil, err
	}

	// go over all measurement files and append measurements to target file
	for _, x := range data {

		filename := fmt.Sprintf("%s/%s", dataDir, x.Name())

		// todo maybe filter here?

		file, err := os.Open(filename)
		if err != nil {
			log.Println("error here", filename)
			return nil, err
		}
		defer file.Close() // todo
		log.Println("appending measurements from", filename)

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

// ExtractNumber gets a byte slice with text (possibly containing a number)
// and tries to extract and parse that number
func ExtractNumber(input []byte) (int, error) {
	re := regexp.MustCompile(`-?\d+`)
	match := re.Find(input)
	numberStr := string(match)
	result, err := strconv.Atoi(numberStr)
	if err != nil {
		return 0, err
	}
	return result, nil
}
