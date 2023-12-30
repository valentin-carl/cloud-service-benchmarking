package consumer

import (
	"consumer/pkg/config"
	"consumer/pkg/utils"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type Consumer struct {
	config *config.Config
}

func NewConsumer(path string) *Consumer {
	return &Consumer{
		config: config.Load(path),
	}
}

func (c *Consumer) Start() {

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

	// create + start workers
	// wait for <experiment-duration>, stop workers
	done := make(chan bool)
	for i := 0; i < c.config.Consumer.NWorkers; i++ {
		go Consume(channel, queue, c.config, i, done)
	}
	time.Sleep(time.Second * time.Duration(c.config.Experiment.Duration))
	done <- true
}
