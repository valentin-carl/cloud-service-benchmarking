package consumer

import (
	"benchmark/lib/config"
	"benchmark/lib/utils"
	"bufio"
	"consumer/pkg/buffer"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"strconv"
	"time"
)

type BufWorker struct {
	id      string
	writer  *bufio.Writer
	config  *config.Config
	channel *amqp.Channel
	queue   amqp.Queue
}

func NewBufWorker(
	id string,
	config *config.Config,
	connection *amqp.Connection,
	queue amqp.Queue,
) *BufWorker {

	channel, err := connection.Channel()
	utils.Handle(err)

	// plan:
	// - create new dir for this experiment run
	// 		=> ./data/experiment-run-X-node-Y
	// - each worker has one buffered writer -> one file per worker
	//		=> ./data/experiment-run-X-node-Y/experiment-run-X-worker-Z.csv
	// - merge these after run is over and output to outDir
	// 		=> ./out/experiment-run-X-node-Y.csv

	// the nodeId is used to generate unique filenames across different nodes
	nodeId, err := utils.GetNodeId()
	utils.Handle(err)

	file, err := os.Create(fmt.Sprintf("%s/%s-node-%d-%s.csv", config.Experiment.DataDir, config.Experiment.Id, nodeId, id))
	utils.Handle(err)

	size, err := buffer.GetBlockSize()
	utils.Handle(err)

	return &BufWorker{
		id:      id,
		writer:  bufio.NewWriterSize(file, int(size)),
		config:  config,
		channel: channel,
		queue:   queue,
	}
}

func (b *BufWorker) Start(stop <-chan bool) {

	defer b.channel.Close()

	// the nodeId is used to generate unique filenames across different nodes
	nodeId, err := utils.GetNodeId()
	utils.Handle(err)

	// register worker at broker
	options := b.config.Consumer.Options
	messages, err := b.channel.Consume(
		b.queue.Name,
		fmt.Sprintf("consumer-%d-%s", nodeId, b.id),
		options.AutoAck,
		options.Exclusive,
		options.NoLocal,
		options.NoWait,
		options.Args,
	)

	// tested locally: improved average throughput from ~33,000 msg/s to ~35,000 msg/s
	// not a very reliable result but enough to leave this in
	err = b.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)

	// write all buffer contents at the end
	// in case the buffer contains measurements that haven't been persisted yet,
	// they are added to the file
	defer func(writer *bufio.Writer) {
		err := writer.Flush()
		utils.Handle(err)
	}(b.writer)

	// worker:
	// (0) get requests
	// (1) extract timestamps
	// (2) write them to buffered reader
	// (3) acknowledge messages
	// (4) flush writer at the end to not miss any measurements

	// todo check again what happens with the merging and whether that needs to be adjusted
	// theory: just put files into dataDir and the rest already happens in main
	// => just a theory tho
	b.ListenAndStore(messages, stop)

}

func (b *BufWorker) ListenAndStore(messages <-chan amqp.Delivery, stop <-chan bool) {
	for {
		select {
		case msg := <-messages:
			{
				// (1) get producer timestamp
				utils.Handle(msg.Headers.Validate()) // todo comment this out once I know it works
				var t int64
				var err error
				if tprod, ok := msg.Headers["tProducer"]; ok {
					// this panics if the type assertion fails
					// that's ok?
					tprod, _ := tprod.(string)
					t, err = strconv.ParseInt(tprod, 10, 64)
					//log.Println(t)
				} else {
					utils.Handle(errors.New("\"tProducer\" header missing from message"))
				}

				// (2) store timestamps
				n, err := b.writer.WriteString(fmt.Sprintf("%d,%d\n", t, time.Now().UnixMilli()))
				utils.Handle(err)
				if n == 0 {
					utils.Handle(errors.New("nothing written, something went wrong"))
				}

				// (3) send manual acknowledgement
				if !b.config.Consumer.Options.AutoAck {
					utils.Handle(msg.Ack(b.config.Consumer.Options.AckMultiple))
				}
			}
		case <-stop:
			{
				log.Println(b.id, "done")
				goto TheEnd
			}
		}
	}
TheEnd:
	// todo test
	// todo make sure everything's written to disk
	//  to test, just run experiment with enough messages to fill the buffers a couple times
	//  if there are any unacknowledged messages in the queue at the end, there's still a bug
	log.Println("worker done")
}

//
// supervisor time!
// the supervisor listens for interrupt and for when the consumer knows
// that all producers are done and stops all workers accordingly
//

type Supervisor struct {
	interrupt <-chan os.Signal
	done      <-chan bool
	config    *config.Config
	channel   *amqp.Channel
	queue     amqp.Queue
}

func NewSupervisor(
	interrupt <-chan os.Signal,
	done <-chan bool,
	config *config.Config,
	connection *amqp.Connection,
	queue amqp.Queue,
) *Supervisor {

	channel, err := connection.Channel()
	utils.Handle(err)

	return &Supervisor{
		interrupt: interrupt,
		done:      done,
		config:    config,
		channel:   channel,
		queue:     queue,
	}
}

// Supervise
// - stop: signal for worker
func (s *Supervisor) Supervise(stop chan<- bool) {

	defer s.channel.Close()

	select {
	case <-s.interrupt:
		// interrupt signal comes form main.go
		{
			// stop the workers immediately (but still flush the buffers!)
			log.Println("supervisor received interrupt, stopping workers immediately")
			close(stop)
		}
	case <-s.done:
		// done comes from consumer
		// means that all producers have sent "end" to consumer
		// but there could still be messages in the queue
		// => wait until empty and then stop workers
		{
			// fixme while waiting for the queue to be empty, an interrupt could be triggered => ignore this for now
			log.Println("supervisor received done message, waiting for queue to be empty before stopping workers")
			t, tmax := 100, 6400

			// rabbitmq wants 'x-quorum-initial-group-size' to be an int but golang parses all numbers in json files as float64
			var args amqp.Table
			if _, ok := s.config.Broker.Queue.Args["x-quorum-initial-group-size"]; !ok {
				args = s.config.Broker.Queue.Args
			} else {
				log.Println("trying to cast x-quorum-initial-group-size to int")
				args = make(amqp.Table) // golang why are maps nil by default????
				for key, value := range s.config.Broker.Queue.Args {
					if key != "x-quorum-initial-group-size" {
						args[key] = value
					} else {
						log.Println("casting")
						args["x-quorum-initial-group-size"] = int(s.config.Broker.Queue.Args["x-quorum-initial-group-size"].(float64))
					}
				}
			}

		Check:
			for {
				// check amount of messages in the queue
				q, err := s.channel.QueueDeclarePassive(
					s.queue.Name,
					s.config.Broker.Queue.Durable,
					s.config.Broker.Queue.AutoDelete,
					s.config.Broker.Queue.Exclusive,
					s.config.Broker.Queue.NoWait,
					args,
				)
				utils.Handle(err)

				log.Println("messages left in queue:", q.Messages)

				// if the queue's empty, close "done" to stop workers
				if q.Messages == 0 {
					// stop the worker by closing the done channel
					log.Println("stopping worker ...")
					close(stop)
					break Check
				} else {
					// if not empty, wait before checking again
					if t < tmax {
						t *= 2
					}
					log.Println(q.Messages, "messages remaining in the queue")
					time.Sleep(time.Duration(t) * time.Millisecond)
					continue Check
				}
			}
		}
	}
	log.Println("supervisor done")
}
