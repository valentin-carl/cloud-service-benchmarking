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
	"sync"
	"time"
)

const (
	dir           = "./data"
	measurementId = 0
)

func main() {

	// create a new directory to store data in if it doesn't exist yet
	err := os.Mkdir(dir, os.ModePerm)
	if err != nil {
		log.Println("dir", dir, "already exists")
	}

	// todo to find number for new files: look at deprecated lib code
	// todo also need a node id in the filename
	// todo how to get that data from the vm

	// new files for each kind of measurement + experiment run
	cpuFile, err := os.Create(fmt.Sprintf("%s/cpu-%d.csv", dir, measurementId))
	utils.Handle(err)

	memFile, err := os.Create(fmt.Sprintf("%s/mem-%d.csv", dir, measurementId))
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
