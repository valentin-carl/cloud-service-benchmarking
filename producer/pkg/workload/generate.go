package workload

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path"
	"producer/pkg/config"
	"regexp"
	"strconv"
)

const (
	a           = 97
	z           = 122
	WorkloadDir = "./workloads/"
)

type Generator struct {
	MessageSize uint
	NMessages   uint
	config      *config.Config
}

func NewGenerator(messageSize, nMessages uint, config *config.Config) *Generator {
	return &Generator{
		MessageSize: messageSize,
		NMessages:   nMessages,
		config:      config,
	}
}

// GenerateMessage generates a single message
func (g *Generator) GenerateMessage() []byte {
	msg := make([]byte, g.MessageSize)
	for i := 0; i < len(msg); i++ {
		msg[i] = byte(rand.Intn(z-a+1) + a)
	}
	return msg
}

// GenerateMessages generates multiple messages
func (g *Generator) GenerateMessages() [][]byte {
	msgs := make([][]byte, g.NMessages)
	for i := 0; i < len(msgs); i++ {
		msgs[i] = g.GenerateMessage()
	}
	return msgs
}

// Store will create a subdirectory of <WorkloadDir> and store a set of messages
// The set of messages will be split into <nSplit> different files
// The idea is to have one workload file per worker, and store/load the workloads
// in the same way to make the experiment as repeatable as possible
func (g *Generator) Store(msgs [][]byte, subdir string, nSplit int) error {

	// validate inputs
	if nSplit <= 0 {
		return errors.New("need input > 0")
	}
	if len(msgs)%nSplit != 0 {
		return errors.New("unable to split workload equally")
	}

	// check if subdir exists and return error if it does
	p := path.Join(WorkloadDir, subdir)
	if _, err := os.ReadDir(p); err == nil {
		return errors.New("directory already exists: " + p)
	}

	// create subdir to store messages
	log.Println(p)
	err := os.Mkdir(p, os.ModePerm)
	if err != nil {
		log.Println("unable to create directory to store workload")
		return err
	}

	// store workload in nSplit files
	nPerFile := len(msgs) / nSplit
	msgIndex, fileIndex := 0, 0

	for ; fileIndex < nSplit; fileIndex++ {

		// create file
		file, err := os.Create(path.Join(p, fmt.Sprintf("workload-worker-%d", fileIndex)))
		if err != nil {
			log.Println("error creating file for index", fileIndex)
			return err
		}

		// write to file
		for i := 0; i < nPerFile; i++ {
			msg := append(msgs[msgIndex], 10)
			n, err := file.Write(msg)
			if err != nil {
				log.Println("error writing msg to file", file.Name())
				return err
			}
			log.Println("wrote", n, "bytes to", file.Name())
			msgIndex++
		}

		err = file.Close()
		if err != nil {
			log.Println("error closing file", file.Name())
			return err
		}
	}

	return nil
}

// GetWorkloadName looks at the config file to find a good dir name to store workload in
func (g *Generator) GetWorkloadName() (name string, err error) {

	var id int

	// find current experiment run number
	matches := regexp.MustCompile(`experiment-run-(\d+)`).FindStringSubmatch(g.config.Experiment.Id)
	if len(matches) >= 2 {
		log.Println(matches)
		id, err = strconv.Atoi(matches[1])
		if err != nil {
			log.Println("error converting match from string to int", matches[1])
			return
		}
	} else {
		log.Println("no match found")
		err = errors.New("no match found")
		return
	}

	// generate workload name
	name = fmt.Sprintf("worldload-run-%d", id)
	log.Printf("generated workload name: '%s'\n", name)
	return
}
