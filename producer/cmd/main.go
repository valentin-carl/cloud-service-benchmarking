package main

import (
	"benchmark/lib/config"
	"benchmark/lib/utils"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	prod "producer/pkg/producer"
	"producer/pkg/workload"
	"time"
)

const (
	configPath = "../config.json"
)

func main() {

	// load config
	c := config.Load(configPath)
	fmt.Println(c.String())

	// the producer needs the stop channel when generating messages on the fly
	// the workloadPath is only necessary when we generate all messages before
	// starting the experiment run
	var (
		stop         chan bool
		workloadPath string
		producer     *prod.Producer
	)

	// to chancel the experiment early
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)

	// generate the workload during the experiment run
	if c.Workload.GenerateRealTime {

		// start the producer
		stop = make(chan bool)
		producer = prod.NewProducer(c)
		producer.Start("", interrupt, stop)

		// let the main goroutine wait here for the duration of the experiment
		timer := time.NewTimer(time.Duration(c.Experiment.Duration) * time.Second) // todo set this correctly in the config
		select {
		case <-timer.C:
			log.Println("timer over")
			goto Done
		case <-interrupt:
			log.Println("interrupted")
			goto Done
		}

		// closing the stop channel stops the generator, which stops the workers
		// once all generators & workers are done, the producer also stops
	Done:
		log.Println("experiment duration over, stopping producer, generators, and workers ...")
		close(stop)
		log.Println("stop channel closed")

	} else {

		// generate the workload now, store it, and load it for the workers later
		// maybe generate new workload
		if c.Workload.Generate && c.Workload.WorkloadPath == "" {

			// generate new wl
			log.Println("generating new workload")
			generator := workload.NewGenerator(c.Experiment.MessageSize, c.Experiment.NMessagesTotal, c)
			wl := generator.GenerateMessages()
			wlName, err := generator.GetWorkloadName()
			utils.Handle(err)
			err = generator.Store(wl, wlName, c.Producer.NWorkers)
			utils.Handle(err)
			workloadPath = wlName

		} else if !c.Workload.Generate && c.Workload.WorkloadPath != "" {

			// load existing one
			// => producer loads it on its own
			log.Println("using existing workload", c.Workload.WorkloadPath)
			workloadPath = c.Workload.WorkloadPath

		} else {

			// not generating a new workload and not specifying a new one won't work
			utils.Handle(errors.New("possible workload-misconfiguration detected"))
		}

		// start the producer + experiment run
		producer = prod.NewProducer(c)
		// 'stop == nil' means the workload is already generated
		// note: this is not running in a new goroutine
		// => experiment is over once the Start function is returns
		producer.Start(workloadPath, interrupt, nil)
	}

	// experiment is over :-)
	log.Println("all done")
}
