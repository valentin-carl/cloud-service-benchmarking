package producer

import (
	"benchmark/lib/config"
	"benchmark/lib/utils"
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

type Worker struct {
	workerId string
	config   *config.Config
	channel  *amqp.Channel
}

func NewWorker(workerId string, config *config.Config, connection *amqp.Connection) *Worker {

	// every worker has a new channel, but they share one connection
	// channels aren't thread safe https://stackoverflow.com/questions/47888730
	// todo adjust consumer accordingly => issue #6
	channel, err := connection.Channel()
	utils.Handle(err)

	return &Worker{
		workerId,
		config,
		channel,
	}
}

// Start
// - workload: get messages one by one from the producer, stop once channel closes
// - ack: tell producer that this worker is done, do so once workload channel is closed
// => producer sends "end" to consumer once all worker acknowledged that they're done
func (w *Worker) Start(channel *amqp.Channel, queue amqp.Queue, messages <-chan []byte, ack chan<- bool) {

	// todo remove later on
	messageCounter := 0

	// the worker is done once the producer closes this channel and all messages are taken from it
	for message := range messages {

		log.Printf("%s sending message %d to consumer\n", w.workerId, messageCounter)

		// create publishing to send message and current timestamp
		publishing := amqp.Publishing{
			Body: message,
		}
		headers := make(amqp.Table)
		headers["tProducer"] = time.Now().UnixMilli()
		publishing.Headers = headers

		// send to broker
		err := channel.PublishWithContext(
			context.Background(),
			"", // publish to default exchange, routing key == queue name gets message to the consumer
			queue.Name,
			w.config.Producer.Options.Mandatory,
			w.config.Producer.Options.Immediate,
			publishing,
		)
		utils.Handle(err)

		// todo remove later on
		messageCounter++
	}

	// reached once the messages channel is closed
	log.Println(w.workerId, "received closed channel, stopping")

	// tell consumer this worker is done
	ack <- true
}

func (w *Worker) Close() error {
	log.Println(w.workerId, "was closed")
	return w.channel.Close()
}
