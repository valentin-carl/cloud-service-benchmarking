package consumer

import (
	"consumer/pkg/config"
	"consumer/pkg/utils"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
)

type Consumer struct {
	config *config.Config
}

func NewConsumer(path string) *Consumer {
	return &Consumer{
		config: config.Load(path),
	}
}

func (c *Consumer) Start(interrupt <-chan os.Signal) {

	// connect to the broker
	connection, err := amqp.Dial(c.config.Broker.URL)
	utils.Handle(err)

	// create a channel to access api
	channel, err := connection.Channel()
	utils.Handle(err)

	// create the queue if it doesn't exist yet
	qc := c.config.Broker.Queue
	queue, err := channel.QueueDeclare(
		qc.Name,
		qc.Durable,
		qc.AutoDelete,
		qc.Durable,
		qc.NoWait,
		qc.Args,
	)
	utils.Handle(err)

	// get channel with "end" message (go channel not amqp channel)
	// context: after a producer is done, it sends an "end" message
	// count and wait until all producers did so, then flush buffer and terminate
	lastMsgChannel, err := c.GetLastMsgChannel(channel)

	// create + start workers
	stop, ack := make(chan bool), make(chan bool)
	workers := make([]*Worker, c.config.Consumer.NWorkers)
	for i := 0; i < len(workers); i++ {
		workers[i] = NewWorker(
			fmt.Sprintf("worker-%d", i),
			c.config,
		)
	}
	for _, worker := range workers {
		go worker.Start(channel, queue, stop, ack)
		log.Printf("worker \"%s\" started\n", worker.workerId)
	}

	// wait for interrupt or signals from producers to stop the workers
	stopAndWaitForWorkers := func() {
		log.Println("consumer telling workers to flush buffers and stop")
		close(stop)
		for i := 0; i < c.config.Consumer.NWorkers; i++ {
			<-ack
		}
		log.Println("all workers acknowledged stopping")
	}
	nProducersDone, prodTotal := 0, c.config.Producer.NProducers
	for {
		select {
		case msg := <-lastMsgChannel:
			{
				if string(msg.Body) == "end" {
					nProducersDone++
					log.Printf("received \"end\" from producer, %d/%d producers are done\n", nProducersDone, prodTotal)
					if nProducersDone == prodTotal {
						log.Println("all producers are done, stopping workers")
						goto Done
					}
				} else {
					log.Println("consumer got message != \"end\" through lastMsgChannel:", string(msg.Body))
				}
			}
		case <-interrupt:
			{
				log.Println("consumer was interrupted, stopping workers")
				goto Done
			}
		}
	}
Done:
	stopAndWaitForWorkers()
}

// GetLastMsgChannel returns a go channel the producers use to signal the end of the experiment
// for reference (this function is more or less taken for there): https://www.rabbitmq.com/tutorials/tutorial-three-go.html
// This uses the fan-out pattern — every producer sends an 'end' message to each consumer
// Once all of them (i.e. all "end messages", same number as number of producers) have been read, the experiment is done
// The config for these queues won't change; hence, it doesn't need to be in the config file
// TODO test with multiple consumer nodes
func (c *Consumer) GetLastMsgChannel(ch *amqp.Channel) (<-chan amqp.Delivery, error) {

	exchange := "lastMsg"

	// exchange
	err := ch.ExchangeDeclare(
		exchange,
		"fanout",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// new tmp queue
	queue, err := ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// bind queue to exchange where producers send "end" messages
	err = ch.QueueBind(
		queue.Name,
		"",
		exchange,
		false,
		nil,
	)

	// register as consumer + return
	return ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
}
