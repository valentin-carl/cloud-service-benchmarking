package producer

import (
	"benchmark/lib/config"
	"benchmark/lib/utils"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
)

type Producer struct {
	config *config.Config
}

func NewProducer(config *config.Config) *Producer {
	return &Producer{
		config: config,
	}
}

func (p *Producer) Start(interrupt <-chan os.Signal) {

	// connect to broker
	connection, err := amqp.Dial(p.config.Broker.URL)
	utils.Handle(err)

	// access api
	channel, err := connection.Channel()
	utils.Handle(err)

	// create queue if it doesn't exist
	queue, err := channel.QueueDeclare()
}
