package workload

import (
	"bufio"
	"errors"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
)

type Workload [][]byte

// LoadWorkloads takes the messages from a subdirectory of /workloads/ and loads
// them into a workload object. The order and workerIds are persevered across
// experiment runs, which allows us to repeat the experiment with the same workload
// and to ensure that the workers get their messages in the same order etc. Caveat:
// we can't really reproduce the OS's and the go-runtime's scheduling  decisions,
// but their influence should not be too large (knock on wood!).
func LoadWorkloads(messageSize uint, dir string) ([]Workload, error) {

	// check if dir exists
	p := path.Join(WorkloadDir, dir)
	allFiles, err := os.ReadDir(p)
	if err != nil {
		log.Println(p, "doesn't exist")
		return nil, err
	}

	// get workload files
	// there might be other stuff in the directory
	wlfiles := make([]os.DirEntry, 0)
	for _, file := range allFiles {

		// check it's a workload file, otherwise log and skip
		log.Println(file.Name())
		if regexp.MustCompile(`workload-worker-\d+`).MatchString(file.Name()) {
			log.Println("found workload file", file.Name())
			wlfiles = append(wlfiles, file)
		}
	}

	// read workloads from files into slices
	workloads := make([]Workload, len(wlfiles))
	for i, file := range wlfiles {

		// get worker number to find slice index
		matches := regexp.MustCompile(`workload-worker-(\d+)`).FindStringSubmatch(file.Name())
		if len(matches) < 2 {
			log.Println("did not find match", file.Name())
			return nil, errors.New("no match found")
		}
		workerId, err := strconv.Atoi(matches[1])
		if err != nil {
			log.Println("couldn't convert match from string to int", matches)
			return nil, err
		}
		log.Println("workerId", workerId)

		// store workload
		workload := make(Workload, 0)
		currentWl, err := os.Open(path.Join(WorkloadDir, dir, file.Name()))
		defer currentWl.Close()
		if err != nil {
			log.Println("could not open file", file.Name())
			return nil, err
		}
		scanner := bufio.NewScanner(currentWl)
		for scanner.Scan() {
			log.Println(len(scanner.Bytes()), string(scanner.Bytes()))
			msg := scanner.Bytes()[:messageSize] // remove trailing newline character
			workload = append(workload, msg)
		}
		workloads[i] = workload
	}

	// return
	return workloads, nil
}
