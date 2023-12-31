package consumer

import (
	buffer "consumer/pkg/buffer"
	"consumer/pkg/config"
	"consumer/pkg/utils"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

type Worker struct {
	workerId string
	buffer   *buffer.Buffer
	config   *config.Config
}

func NewWorker(workerId string, bufferSize uint, config *config.Config) *Worker {
	return &Worker{
		workerId: workerId,
		buffer:   buffer.NewBuffer(bufferSize),
		config:   config,
	}
}

func (w *Worker) Consume(channel *amqp.Channel, queue amqp.Queue, done <-chan bool) {

	// flush the buffer at the end
	defer func(buffer *buffer.Buffer, workerId string, config *config.Config) {
		utils.Handle(buffer.Close(workerId, config))
	}(w.buffer, w.workerId, w.config)

	// register as consumer at broker
	options := w.config.Consumer.Options
	consumerStr := fmt.Sprintf("consumer-%d-%s", w.config.Consumer.Node, w.workerId)
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
				// read producer timestamp from event header
				log.Println(w.workerId, "received message")
				err := event.Headers.Validate()
				utils.Handle(err)
				tProd, ok := event.Headers["tProducer"]
				if !ok {
					utils.Handle(errors.New("could not read tProducer header"))
				}
				tProducer, ok := tProd.(int64)
				if !ok {
					utils.Handle(errors.New("could not transform tProducer to int64"))
				}

				// store measurement in buffer
				err = w.buffer.Store(
					w.workerId,
					w.config,
					buffer.Measurement{
						TProducer: tProducer,
						TConsumer: time.Now().UnixMilli(),
					},
				)
				utils.Handle(err)

				// (depending on configuration) send acknowledgement to broker
				if !options.AutoAck {
					err := event.Ack(options.AckMultiple)
					utils.Handle(err)
				}
			}
		case <-done:
			{
				// todo remember that break here will (I think?) only break the select and not the for loop
				//  => check csb-temp
				log.Println("worker", w.workerId, "is done, flushing buffer")
				goto ClockOff
			}
		}
	}
ClockOff:
	log.Println("worker", w.workerId, "is going home")
}
