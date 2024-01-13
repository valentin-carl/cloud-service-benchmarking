package producer

import (
	"benchmark/lib/config"
	"benchmark/lib/utils"
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"strconv"
	"time"
)

type Worker struct {
	workerId string
	messages <-chan Message
	config   *config.Config
	channel  *amqp.Channel
	queue    amqp.Queue
}

func NewWorker(workerId string, messages <-chan Message, config *config.Config, connection *amqp.Connection, queue amqp.Queue) *Worker {

	// every worker has a new channel, but they share one connection
	// channels aren't thread safe https://stackoverflow.com/questions/47888730
	// todo adjust consumer accordingly => issue #6
	channel, err := connection.Channel()
	utils.Handle(err)

	return &Worker{
		workerId,
		messages,
		config,
		channel,
		queue,
	}
}

// Start
// - workload: get messages one by one from the producer, stop once channel closes
// - ack: tell producer that this worker is done, do so once workload channel is closed
// => producer sends "end" to consumer once all worker acknowledged that they're done
func (w *Worker) Start() {
	defer w.channel.Close()
	for msg := range w.messages {
		log.Println(w.workerId, "sending message..")
		headers := make(amqp.Table)
		tprodstr := strconv.FormatInt(time.Now().UnixMilli(), 10)
		headers["tProducer"] = tprodstr
		// add timestamp as header 'tProducer'
		// => this avoids having to use an external plugin that isn't precise enough
		pub := amqp.Publishing{
			Headers: headers,
			Body:    msg,
		}
		err := w.channel.PublishWithContext(
			context.Background(),
			"",           // "" means message is sent to default exchange
			w.queue.Name, // routing key == queue name gets message to the right place
			false,
			false,
			pub, // actual message
		)
		utils.Handle(err)
	}
	log.Printf("%s: messages channel closed, stopping..\n", w.workerId)
}
