package producer

import (
	"benchmark/lib/config"
	"benchmark/lib/utils"
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"producer/pkg/workload"
	"sync"
)

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
func (p *Producer) Start(workloadPath string, interrupt <-chan os.Signal) {

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

	// load workload to distribute
	workloads, err := workload.LoadWorkloads(workloadPath)
	utils.Handle(err)

	// create workers
	msgChannels := make([]chan []byte, p.config.Producer.NWorkers)
	workers := make([]*Worker, p.config.Producer.NWorkers)
	for i := 0; i < p.config.Producer.NWorkers; i++ {
		msgChannels[i] = make(chan []byte)
		workerId := fmt.Sprintf("worker-%d", i)
		workers[i] = NewWorker(workerId, msgChannels[i], p.config, connection, queue)
		log.Println("created", workerId)
	}

	// start all workers and tell them which messages to send
	// todo also use waitgroup in consumer => see issue on github
	var wg sync.WaitGroup
	for i := 0; i < p.config.Producer.NWorkers; i++ {
		go workers[i].Start()
		go DistributeWorkload(workloads[i], msgChannels[i], interrupt, &wg)
	}
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
func DistributeWorkload(workload workload.Workload, messages chan<- []byte, interrupt <-chan os.Signal, wg *sync.WaitGroup) {

	log.Println("distributing workload..")

	wg.Add(1)
	defer wg.Done()
	defer close(messages) // tell workers to stop

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
