package producer

import (
	"benchmark/lib/config"
	"benchmark/lib/utils"
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"math/rand"
	"os"
	"producer/pkg/workload"
	"sync"
	"time"
)

const (
	a = 97
	z = 122
)

type Message []byte

type Producer struct {
	config *config.Config
}

func NewProducer(config *config.Config) *Producer {
	return &Producer{
		config: config,
	}
}

// Start creates worker, loads messages, and tells the workers to send them to the broker
// - pass workload path as input, either from config if pre-existing, or from where it was stored earlier
// - stop: channel that can be used to stop an experiment
// for now: if stop == nil, load the generated messages
// if stop != nil, generate them on the fly and stop when stop channel closes
func (p *Producer) Start(workloadPath string, interrupt <-chan os.Signal, stop <-chan bool) {

	// connect to broker
	connection, err := amqp.Dial(p.config.Broker.URL)
	utils.Handle(err)
	defer connection.Close()

	// access api
	channel, err := connection.Channel()
	utils.Handle(err)
	defer channel.Close()

	// create queue if it doesn't exist
	qConfig := p.config.Broker.Queue
	queue, err := channel.QueueDeclare(
		qConfig.Name,
		qConfig.Durable,
		qConfig.AutoDelete,
		qConfig.Exclusive,
		qConfig.NoWait,
		qConfig.Args,
	)
	utils.Handle(err)

	// create workers + channels to send them messages through
	msgChannels := make([]chan Message, p.config.Producer.NWorkers)
	workers := make([]*Worker, p.config.Producer.NWorkers)
	for i := 0; i < p.config.Producer.NWorkers; i++ {
		msgChannels[i] = make(chan Message)
		workerId := fmt.Sprintf("worker-%d", i)
		workers[i] = NewWorker(workerId, msgChannels[i], p.config, connection, queue)
		log.Println("created", workerId)
	}

	// start all workers and tell them which messages to send
	var wg sync.WaitGroup
	wg.Add(p.config.Producer.NWorkers) // note: NWorkers == NGenerators

	// depending on the config, the workload has already been generated and needs to be loaded,
	// or it will be generated in real time during the experiment
	generateInRealTime := stop != nil
	if generateInRealTime {
		// start a new real time generator for each worker and send the messages to the workers
		for i := 0; i < p.config.Producer.NWorkers; i++ {
			go workers[i].Start()
			i := i
			go func() {
				var RTGseed int64 = int64(i)
				NewRTG(RTGseed, p.config.Experiment.MessageSize, stop).Generate(msgChannels[i]) // stopped in main.go
				wg.Done()
			}()
		}

	} else {

		// load workloads to distribute
		workloads, err := workload.LoadWorkloads(workloadPath)
		utils.Handle(err)

		// start the workers and send the loaded workloads to the workers
		for i := 0; i < p.config.Producer.NWorkers; i++ {
			go workers[i].Start()
			i := i
			go func() {
				DistributeWorkload(workloads[i], msgChannels[i], interrupt) // stops on its own when all messages are sent
				wg.Done()
			}()
		}
	}

	// wait here until all messages are sent
	// (or, when generating in real time, until the generators are stopped by
	// closing the 'stop' channel in the main goroutine
	wg.Wait()
	log.Println("wait group all done")

	// send "end" message to lastMsg exchange with fan-out
	// => tells all consumers that this producer is done,
	// they wait for all to be done before stopping
	err = p.sendLastMsg(channel)
	utils.Handle(err)
	log.Println("send \"end\" to all consumers")
}

// Sends messages to one worker to send
func DistributeWorkload(workload workload.Workload, messages chan<- Message, interrupt <-chan os.Signal) {

	log.Println("distributing workload..")
	defer close(messages) // tell workers to stop
	time.Sleep(5 * time.Second)

	for i, msg := range workload {
		select {
		case messages <- msg:
			log.Println("message", i, "sent")
		case <-interrupt:
			return
		}
	}
}

func (p *Producer) sendLastMsg(channel *amqp.Channel) error {

	const (
		exchange = "lastMsg"
		lastMsg  = "end"
	)

	// declare the exchange
	err := channel.ExchangeDeclare(
		exchange,
		"fanout",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// send the message
	return channel.PublishWithContext(
		context.Background(),
		exchange, // exchange name is sufficient
		"",       // each consumer waits on a different queue, hence no routing key + fanout
		false,
		false,
		amqp.Publishing{
			Body: []byte(lastMsg),
		},
	)
}

//
// code to generate messages on the fly
//

// RTG == real time generator
type RTG struct {
	seed int64
	size uint
	stop <-chan bool
}

func NewRTG(seed int64, size uint, stop <-chan bool) *RTG {
	return &RTG{
		seed: seed, // for now, the seed is the worker id => seed for whole experiment: nWorkers
		size: size,
		stop: stop,
	}
}

// use as `go g.Generate(...)`
func (g *RTG) Generate(messages chan<- Message) {

	// stop the worker
	defer close(messages)

	// generate random message
	s := rand.NewSource(g.seed)
	r := rand.New(s)
	generate := func() Message {
		msg := make(Message, g.size)
		for i := 0; i < int(g.size); i++ {
			msg[i] = byte(r.Intn(z+1-a) + a)
		}
		return msg
	}

	// send message to worker
	for {
		m := generate()
		select {
		case messages <- m:
			{
				log.Println("generated message:", string(m))
			}
		case <-g.stop:
			{
				log.Println("stop channel closed, stopping ...")
				goto TheEnd
			}
		}
	}
TheEnd:
	log.Println("generator done")
}
