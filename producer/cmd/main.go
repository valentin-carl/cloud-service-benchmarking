package main

import (
	"benchmark/lib/config"
	"benchmark/lib/utils"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"producer/pkg/producer"
	"producer/pkg/workload"
)

const (
	configPath = "../config.json"
)

// todo for experiment: reuse same workload 3 times

func main() {

	// verify config can be loaded
	c := config.Load("../config.json")
	fmt.Println(c.String())

	// maybe generate new workload
	var workloadPath string
	if c.Workload.Generate && c.Workload.WorkloadPath == "" {
		// generate new wl
		log.Println("generating new workload")
		generator := workload.NewGenerator(c.Experiment.MessageSize, c.Experiment.NMessagesTotal, c)
		wl := generator.GenerateMessages()
		wlName, err := generator.GetWorkloadName()
		utils.Handle(err)
		err = generator.Store(wl, wlName, c.Producer.NWorkers)
		utils.Handle(err)
		//workloadPath = fmt.Sprintf("workloads/%s", wlName) // todo test
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

	log.Println(workloadPath)
	// todo test until here

	// create new producer
	log.Println("workloadPath:", workloadPath) // todo test for generated + loaded
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)
	producer := producer.NewProducer(c)
	log.Println("starting producer")
	producer.Start(workloadPath, interrupt)

	// experiment is over :-)
	log.Println("all done")
}
