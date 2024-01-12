package consumer

import (
	"benchmark/lib/config"
	"benchmark/lib/utils"
	buffer "consumer/pkg/buffer"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"strconv"
	"time"
)

type Worker struct {
	workerId string
	buffer   *buffer.Buffer
	config   *config.Config
	channel  *amqp.Channel
	queue    amqp.Queue
}

func NewWorker(connection *amqp.Connection, workerId string, config *config.Config, queue amqp.Queue) *Worker {

	// crate a new channel for each worker because they're not thread-safe!
	channel, err := connection.Channel()
	utils.Handle(err)

	return &Worker{
		workerId: workerId,
		buffer:   buffer.NewBuffer(buffer.CalcOptimalBufferSize()),
		config:   config,
		channel:  channel,
		queue:    queue,
	}
}

// Start todo docs
// - stop: consumer tells workers to stop
// - ack: worker tells consumer that it's really done => important for writing and moving files
func (w *Worker) Start(stop <-chan bool) {

	// get nodeId
	var nodeId int
	nidStr := os.Getenv("NODEID")
	if nidStr == "" {
		log.Panic("nodeId not set, terminating ...")
	} else {
		var err error // same problem as in main
		nodeId, err = strconv.Atoi(nidStr)
		utils.Handle(err)
		log.Printf("nodeId set to %d\n", nodeId)
	}

	// register as consumer at broker
	options := w.config.Consumer.Options
	consumerStr := fmt.Sprintf("consumer-%d-%s", nodeId, w.workerId)
	events, err := w.channel.Consume(
		w.queue.Name,
		consumerStr,
		options.AutoAck,
		options.Exclusive,
		options.NoLocal,
		options.NoWait,
		options.Args,
	)
	utils.Handle(err)
	log.Printf("worker \"%s\" successfully registered as consumer at broker, starting to record measurements\n", w.workerId)

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
				log.Println(tProd)
				if !ok {
					utils.Handle(errors.New("could not read tProducer header"))
				}
				//tProducerStr, ok := tProd.(string)
				tProducer, ok := tProd.(int64)
				if !ok {
					utils.Handle(errors.New("could not transform tProducer to int64"))
				}
				//tProducer, err := strconv.Atoi(tProducerStr)
				utils.Handle(err)

				// store measurement in buffer
				err = w.buffer.Store(
					w.workerId,
					w.config,
					buffer.Measurement{
						//TProducer: int64(tProducer),
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
		case <-stop:
			{
				log.Println("worker", w.workerId, "is done, flushing buffer")
				goto ClockOff
			}
		}
	}
ClockOff:
	// flush the buffer at the end
	utils.Handle(w.buffer.Close(w.workerId, w.config))
	log.Println("worker", w.workerId, "is going home")
}
