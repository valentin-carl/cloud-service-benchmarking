package main

import (
	"consumer/pkg/config"
	"consumer/pkg/consumer"
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
		nodeId, err := strconv.Atoi(nidStr)
		utils.Handle(err)
		log.Printf("nodeId set to %d\n", nodeId)
	}

	// load config
	conf := config.Load(configFile)
	err := utils.ArchiveMeasurements(conf.Experiment.DataDir)
	utils.Handle(err)

	// create consumer
	cons := consumer.NewConsumer(configFile)

	// start workers
	// ensure data is stored in case of sigint
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	cons.Start(c)

	// merge all measurements into a single csv file containing all measurements
	targetFile := conf.Experiment.Id + "-" + strconv.Itoa(nodeId) + ".csv"
	_, err = utils.MergeMeasurements(targetFile, conf.Experiment.DataDir, conf.Experiment.OutDir)
	utils.Handle(err)

	// archive raw data and update config file with new experiment number
	err = utils.ArchiveMeasurements(conf.Experiment.DataDir)
	utils.Handle(err)
	err = conf.IncrementExperimentId(configFile)
	utils.Handle(err)

	log.Printf("the end :-)")
}

/*
Notizen
- channel von worker -> consumer, damit der wartet?
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
