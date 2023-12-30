package consumer

import (
	"consumer/pkg/config"
	"consumer/pkg/utils"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

type Worker struct {
	workerId string
	buffer   *Buffer
	config   *config.Config
}

func NewWorker(workerId string, bufferSize uint, config *config.Config) *Worker {
	return &Worker{
		workerId: workerId,
		buffer:   NewBuffer(bufferSize),
		config:   config,
	}
}

func (w *Worker) Consume(channel *amqp.Channel, queue amqp.Queue, workerId int, done <-chan bool) {

	// register as consumer at broker
	options := w.config.Consumer.Options
	consumerStr := fmt.Sprintf("consumer-%d-%d", w.config.Consumer.Node, workerId)
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
				var measurement struct {
					tProducer int64
					tConsumer int64
				}
				// todo get measurement

				// store measurement in buffer
				err := w.buffer.Store(w.workerId, w.config, measurement)
				utils.Handle(err)

				// (depending on configuration) send acknowledgement to broker
				if !options.AutoAck {
					err := event.Ack(options.AckMultiple)
					utils.Handle(err)
				}
			}
		case <-done:
			{
				log.Println("worker", workerId, "is done")
				// todo write remaining buffer contents to disk

				// todo remember that break here will (I think???) only break the select and not the for loop
				//  => check csb-temp

				goto ClockOff
			}
		}
	}
ClockOff:
	log.Println("worker", workerId, "is going home")
}
