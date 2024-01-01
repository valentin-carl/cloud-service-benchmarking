package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"producer/pkg/utils"
	"regexp"
	"strconv"
)

// Config is similar for consumers & producers => only one file!
// hence, there are some options here that aren't relevant for the consumer
type Config struct {
	Broker struct {
		URL   string `json:"url"`
		Queue struct {
			Name       string         `json:"name"`
			Durable    bool           `json:"durable"`
			AutoDelete bool           `json:"autoDelete"`
			Exclusive  bool           `json:"exclusive"`
			NoWait     bool           `json:"noWait"`
			Args       map[string]any `json:"args"`
		} `json:"queue"`
	} `json:"broker"`

	Producer struct {
		NProducers int `json:"nProducers"`
	} `json:"producer"`

	Consumer struct {
		NWorkers int `json:"nWorkers"`
		Options  struct {
			AutoAck     bool           `json:"autoAck"`
			AckMultiple bool           `json:"ackMultiple"`
			Exclusive   bool           `json:"exclusive"`
			NoLocal     bool           `json:"noLocal"`
			NoWait      bool           `json:"noWait"`
			Args        map[string]any `json:"args"`
		} `json:"options"`
	} `json:"consumer"`

	Workload struct {
		Generate     bool   `json:"generate"`
		WorkloadPath string `json:"workloadPath"`
	} `json:"workload"`

	Experiment struct {
		Id             string `json:"id"`
		DataDir        string `json:"dataDir"`
		OutDir         string `json:"outDir"`
		Duration       int    `json:"duration"`
		NMessagesTotal int    `json:"nMessagesTotal"`
	} `json:"experiment"`
}

func Load(path string) *Config {

	file, err := os.Open(path)
	utils.Handle(err)
	defer file.Close()

	var config Config
	err = json.NewDecoder(file).Decode(&config)
	utils.Handle(err)

	return &config
}

func (c *Config) String() string {
	confStr, err := json.MarshalIndent(c, "", "  ")
	utils.Handle(err)
	return string(confStr)
}

// IncrementExperimentId takes the number from the experiment id in the config file,
// increments it, and writes the new version back to disk
// This is useful because some of the buffer and utils functions rely on the id
// for creating and reading files
func (c *Config) IncrementExperimentId(confPath string) error {

	// get current id
	experimentId := c.Experiment.Id
	match := regexp.MustCompile(`experiment-run-(\d+)`).FindStringSubmatch(experimentId)

	// increment
	if len(match) == 0 {
		return errors.New("did not find number in experiment id")
	}
	num, err := strconv.Atoi(match[1])
	if err != nil {
		return err
	}
	num++
	log.Printf("new experiment id is %d\n", num)

	// write to config file
	c.Experiment.Id = fmt.Sprintf("experiment-run-%d", num)
	configFile, err := os.Create(confPath)
	if err != nil {
		return err
	}
	s := c.String()
	n, err := configFile.Write([]byte(s))
	log.Printf("wrote %d bytes while trying to increment experiment id in config file\n", n)
	return err
}
