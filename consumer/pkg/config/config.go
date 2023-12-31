package config

import (
	"consumer/pkg/utils"
	"encoding/json"
	"os"
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
		NProducers int `json:"NProducers"`
	} `json:"producer"`

	Consumer struct {
		NWorkers   int `json:"nWorkers"`
		BufferSize int `json:"bufferSize"`
		Node       int `json:"node"`
		Options    struct {
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
		Duration       int    `json:"duration"` // TODO change to time.Duration and deal with nanoseconds
		NMessagesTotal int    `json:"NMessagesTotal"`
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
