package main

import (
	"benchmark/lib/utils"
	"fmt"
	"github.com/VividCortex/multitick"
	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	"log"
	"monitoring/pkg/monitor"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"
)

const (
	dir = "./data"
)

func main() {

	// todo think about how to get that data from the vm
	// todo test on linux

	nodeId, err := utils.GetNodeId()
	utils.Handle(err)
	log.Println("got nodeId:", nodeId)

	// create a new directory to store data in if it doesn't exist yet
	err = os.Mkdir(dir, os.ModePerm)
	if err != nil {
		log.Println("dir", dir, "already exists")
	}

	// new files for each kind of measurement + experiment run
	runId, err := GetNextExpNumber(dir)
	utils.Handle(err)

	cpuFile, err := os.Create(fmt.Sprintf("%s/broker-%d-run-%d-cpu.csv", dir, nodeId, runId))
	utils.Handle(err)

	memFile, err := os.Create(fmt.Sprintf("%s/broker-%d-run-%d-memory.csv", dir, nodeId, runId))
	utils.Handle(err)

	// the multitick ticker broadcasts time.Time values to multiple subscribers
	// closing the stop channel broadcasts a signal to all listening goroutines
	// (i.e., to multiple monitor objects)
	ticker := multitick.NewTicker(time.Second, 0)
	stop := make(chan bool)

	// create monitoring routines
	var wg sync.WaitGroup
	cpuMonitor := monitor.NewMonitor[*cpu.Stats](cpuFile)
	memMonitor := monitor.NewMonitor[*memory.Stats](memFile)
	go StartMonitor(cpuMonitor, &wg, ticker.Subscribe(), stop)
	go StartMonitor(memMonitor, &wg, ticker.Subscribe(), stop)

	// listen for interrupt in main and close stop channel accordingly
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println("interrupted")

	// stop the monitors and wait for them to finish
	close(stop)
	wg.Wait()

	log.Println("the end :-)")
}

// Monitor is a helper interface
// all instances of the generic monitor.Monitor implement this interface,
// this makes the Start Function a bit nicer
type Monitor interface {
	Start(<-chan time.Time, <-chan bool)
}

func StartMonitor(m Monitor, wg *sync.WaitGroup, ticker <-chan time.Time, stop <-chan bool) {
	wg.Add(1)
	m.Start(ticker, stop)
	wg.Done()
}

//
// helper functions
//

// GetNextExpNumber goes through the data dir and check which runId to use next
func GetNextExpNumber(dir string) (int, error) {

	// read all files + directories in dataDir
	files, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	// use capture group to find correct experiment number
	// broker-X-run-Y-cpu.csv
	// broker-X-run-Y-memory.csv
	// broker-X-run-Y-network.csv
	var nums []int
	re := regexp.MustCompile(`broker-\d+-run-(\d+)-\D+.csv`)
	for _, file := range files {
		match := re.FindStringSubmatch(file.Name())
		if len(match) == 2 {
			num, err := strconv.Atoi(match[1])
			if err == nil {
				nums = append(nums, num)
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
