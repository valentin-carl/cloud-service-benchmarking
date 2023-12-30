package consumer

import (
	"consumer/pkg/config"
	"consumer/pkg/utils"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

type Worker struct {
	Buffer Buffer
}

func (Worker) Consume(channel *amqp.Channel, queue amqp.Queue, config *config.Config, workerId int, done <-chan bool) {

	// register as consumer at broker
	options := config.Consumer.Options
	consumerStr := fmt.Sprintf("consumer-%d-%d", config.Consumer.Node, workerId)
	events, err := channel.Consume(
		queue.Name,
		consumerStr,
		options.AutoAck,
		options.Exclusive,
		options.NoLocal,
		options.NoWait,
		options.Args,
	)
	utils.Handle(err)

	// read + handle messages
	for {
		select {
		case event := <-events:
			{
				// todo store messages + timestamps

				if !options.AutoAck {
					err := event.Ack(options.AckMultiple)
					utils.Handle(err)
				}
			}
		case <-done:
			{
				log.Println("worker", workerId, "is done.")
				// todo write remaining buffer contents to disk

				// todo remember that break here will (I think???) only break the select and not the for loop
				//  => check csb-temp

				goto ClockOff
			}
		}
	}
ClockOff:
	log.Println("worker", workerId, "done")
}
