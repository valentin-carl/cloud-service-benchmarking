package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
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
	log.Println("found", len(data), "potential measurement files", data)

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
	// write column names to .csv file
	err = write("tProducer, tConsumer")
	if err != nil {
		log.Println("error")
		return nil, err
	}

	// go over all measurement files and append measurements to target file
	log.Println("data:", data, len(data))
	for _, x := range data {

		filename := fmt.Sprintf("%s/%s", dataDir, x.Name())
		log.Println("trying to merge measurements from", filename)

		// todo maybe filter here? => works for now, there shouldn't be any other files here

		if x.IsDir() {
			log.Println(x.Name(), "is a directory, skipping ...")
			continue
		}
		if strings.HasPrefix(x.Name(), ".") {
			log.Println(filename, "is dotfile, skipping ...")
			continue
		}

		file, err := os.Open(filename)
		if err != nil {
			log.Println("could not open", filename)
			return nil, err
		}
		defer file.Close()
		log.Println("appending measurements from", filename)

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			log.Println(filename, line)
			if line == "tProducer, tConsumer" {
				continue
			}
			err = write(line)
			if err != nil {
				log.Println("error writing line", line)
				return nil, err
			}
		}
	}

	// merge successful
	log.Println("merge successful") // x doubt
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

// NewMeasurements looks whether there are any .csv files in a given directory
// Useful to proxy if there are any measurement in the data directory,
// trying to prevent data loss (i.e. accidentally overwriting old measurements) :-)
func NewMeasurements(dir string) (bool, error) {
	// go over all files in directory
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Println(err)
		return false, err
	}
	log.Println(files)
	// check whether file name starts with "experiment-run"
	// initially wanted to checks whether filename ends with ".csv"
	for _, file := range files {
		if !file.IsDir() && regexp.MustCompile("experiment-run").MatchString(file.Name()) {
			return true, nil
		}
	}
	return false, nil
}
