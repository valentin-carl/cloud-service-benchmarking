package main

import (
	"consumer/pkg/config"
	"consumer/pkg/consumer"
	"consumer/pkg/server"
	"consumer/pkg/utils"
	"log"
	"os"
	"os/signal"
	"strconv"
)

const (
	configFile = "./config.json"
)

var (
	nodeId int
)

func main() {

	// read nodeId
	nidStr := os.Getenv("NODEID")
	if nidStr == "" {
		log.Panic("nodeId not set, terminating ...")
	} else {
		var err error // ensures that the global nodeId gets new values and isn't shadowed
		nodeId, err = strconv.Atoi(nidStr)
		utils.Handle(err)
		log.Printf("nodeId set to %d\n", nodeId)
	}

	// load config
	conf := config.Load(configFile)

	// create consumer
	cons := consumer.NewConsumer(configFile)

	// start workers
	// ensure data is stored in case of sigint
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	cons.Start(c)

	// Note: the consumer waits until all workers are completely done
	// -> At this point, all measurements should be written to disk

	// merge all measurements into a single csv file containing all measurements
	// filename pattern "experiment-run-<experiment id>-node-<node id>"
	// (node refers to the consumer node, not broker)
	targetFile := conf.Experiment.Id + "-node-" + strconv.Itoa(nodeId) + ".csv"
	_, err := utils.MergeMeasurements(targetFile, conf.Experiment.DataDir, conf.Experiment.OutDir)
	utils.Handle(err)

	// archive raw data and update config file with new experiment number
	err = utils.ArchiveMeasurements(conf.Experiment.DataDir, conf.Experiment.Id, nodeId)
	utils.Handle(err)

	// note: experiment id is only incremented if this part is reached,
	// i.e. if there were no errors before
	// this could cause inconsistency between consumers if multiple are run
	// TODO fix this if it becomes an issue
	err = conf.IncrementExperimentId(configFile)
	utils.Handle(err)

	// make downloads via http possible => download data from VMs after experiment is done
	// e.g., curl localhost:80/download/out/experiment-run-0-node-0.csv
	// hint: using localhost:80/download in browser allows you to explore all files
	s := server.NewServer(":80", "./")
	utils.Handle(s.Serve())

	log.Printf("the end :-)")
}

/*
Notizen
- config: entweder länge oder anzahl nachrichten vorgeben, damit throughput
	gemessen wird; wenn beides vorgegeben, gebe ich den max throughput ja vor
=> producer bauen der beides kann; dann bei support fragen was besser ist

Producer:
- workload generieren/laden trennen, workload speichern
- beide modi implementieren
A: zeit fest, wieviele nachrichten gehen durch?
B: anzahl nachrichten fest, wie lange dauert das?
- producer schickt am ende eine quit nachricht an eine andere queue?
	die hört sich der consumer an und beendet die worker?
*/

/*
TODO test
	- ob das auch klappt mit vielen nachrichten, wenn die buffer mal voll sind etc.
	- ob das auch klappt wenn es mehrere consumer nodes gibt
*/
