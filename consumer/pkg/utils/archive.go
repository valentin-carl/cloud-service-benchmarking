package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

// GetNextExpNumber looks at all subdirectories in the DataDir (technically any)
// and check how many subdirectories with old data there are
// the subdirectory-names follow the pattern "experiment-run-<experiment number>-<node id>"
// todo remove
func GetNextExpNumber(dir string) (int, error) {

	// read all files + directories in dataDir
	files, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	// use capture group to find correct experiment number
	var nums []int
	re := regexp.MustCompile(`experiment-run-(\d+)-.*`)
	for _, file := range files {
		if file.IsDir() {
			match := re.FindStringSubmatch(file.Name())
			if len(match) == 2 {
				num, err := strconv.Atoi(match[1])
				if err == nil {
					nums = append(nums, num)
				}
			}
		}
	}

	// calculate + return the next experiment number
	if len(nums) == 0 {
		log.Println("did not find any data from old experiments in", dir)
		return 0, nil
	}
	sort.Ints(nums)
	return nums[len(nums)-1] + 1, nil
}

// MoveMeasurements takes all .csv files in a directory and moves them into another
// Used to move all measurements into a different directory
func MoveMeasurements(fromDir string, toDir string) error {

	// get all files from directory
	files, err := os.ReadDir(fromDir)
	if err != nil {
		return err
	}

	// go over all files and move them
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".csv" {
			sourcePath := filepath.Join(fromDir, file.Name())
			destinationPath := filepath.Join(toDir, file.Name())
			err := os.Rename(sourcePath, destinationPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ArchiveMeasurements creates a new dir and moves all measurements
// at top level of dataDir into subdir
func ArchiveMeasurements(dataDir, experimentId string, nodeId int) error {

	// check if this actually has to be run
	if found, err := NewMeasurements(dataDir); err == nil && !found {
		log.Println("no new measurements found, stopping early")
		return nil
	} else if err != nil {
		log.Println("error while looking for new measurements")
		return err
	}

	// create new subdir for most recent experiment measurements
	subdir := fmt.Sprintf("%s/%s-node-%d", dataDir, experimentId, nodeId)
	err := os.Mkdir(subdir, os.ModePerm)
	if err != nil {
		return err
	}
	log.Println("create new dir for measurements")

	// move the data into new subdir
	err = MoveMeasurements(dataDir, subdir)
	log.Println("moved measurements to new dir")
	return err
}
