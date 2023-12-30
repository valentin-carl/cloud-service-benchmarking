package main

import (
	"consumer/pkg/config"
	"consumer/pkg/consumer"
	"consumer/pkg/utils"
	"log"
	"strconv"
)

const (
	configFile = "./config.json"
	nodeId     = 0 // todo find better solution for this => environment variable?
)

func main() {

	// load config
	// todo remove, just load here for debugging
	conf := config.Load(configFile)
	log.Println("loaded configuration", conf.String())

	// create consumer and wait for start signal
	consumer := consumer.NewConsumer(configFile)

	// merge all dumps into a single csv file containing all measurements
	targetFile := conf.Experiment.Id + "-" + strconv.Itoa(nodeId) + ".csv"
	_, err := utils.MergeDumps(targetFile, "./data")
	utils.Handle(err)
}
