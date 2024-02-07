package consumer

import (
	"benchmark/lib/config"
	"benchmark/lib/utils"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"sync"
)

type Consumer struct {
	config *config.Config
}

func NewConsumer(path string) *Consumer {
	return &Consumer{
		config: config.Load(path),
	}
}

// StartWithBufWorkers uses the new and improved workers!
//   - the interrupt channel is _mainly_ used by the supervisor to stop the workers in case
//     of an interrupt, and the consumer just passes it along, but the consumer
//     also stops (after the workers are done) in case of an interrupt
func (c *Consumer) StartWithBufWorkers(interrupt <-chan os.Signal) {

	// connect to broker + create channel to access API
	connection, err := amqp.Dial(c.config.Broker.URL)
	utils.Handle(err)
	defer connection.Close()
	channel, err := connection.Channel()
	utils.Handle(err)
	defer channel.Close()

	// ensure the queue exists before starting supervisor + workers
	qc := c.config.Broker.Queue
	log.Printf("\nqc.Name: %s\nqc.Durable: %t\nqc.AutoDelete: %t\nqc.Exclusive: %t\nqc.NoWait: %t", qc.Name, qc.Durable, qc.AutoDelete, qc.Exclusive, qc.NoWait)

	// rabbitmq wants 'x-quorum-initial-group-size' to be an int but golang parses all numbers in json files as float64
	var args amqp.Table
	if _, ok := qc.Args["x-quorum-initial-group-size"]; !ok {
		args = qc.Args
	} else {
		log.Println("trying to cast x-quorum-initial-group-size to int")
		args = make(amqp.Table) // golang why are maps nil by default????
		for key, value := range qc.Args {
			if key != "x-quorum-initial-group-size" {
				args[key] = value
			} else {
				log.Println("casting")
				args["x-quorum-initial-group-size"] = int(qc.Args["x-quorum-initial-group-size"].(float64))
			}
		}
	}

	queue, err := channel.QueueDeclare(
		qc.Name,
		qc.Durable,
		qc.AutoDelete,
		qc.Exclusive,
		qc.NoWait,
		args,
	)
	utils.Handle(err)

	// get channel with "end" message (go channel not amqp channel)
	// context: after a producer is done, it sends an "end" message
	// count and wait until all producers did so, then flush buffer and terminate
	lastMsgChannel, err := c.GetLastMsgChannel(channel)
	utils.Handle(err)

	// stop channel is used by supervisor to stop the workers
	// create it here because we need to pass it to both
	stop := make(chan bool)

	// done channel is used by consumer to tell the supervisor
	// that all producers are done for this experiment run
	done := make(chan bool)

	// create + start the supervisor
	supervisor := NewSupervisor(interrupt, done, c.config, connection, queue)
	go supervisor.Supervise(stop)

	// create + start the workers
	n := c.config.Consumer.NWorkers
	workers := make([]*BufWorker, n)
	for i := 0; i < n; i++ {
		workers[i] = NewBufWorker(
			fmt.Sprintf("worker-%d", i),
			c.config,
			connection,
			queue,
		)
	}
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		i := i
		go func() {
			wg.Add(1)
			workers[i].Start(stop)
			wg.Done()
		}()
	}

	// wait for and count "end" messages from the producers
	nProducersDone, nProdTotal := 0, c.config.Producer.NProducers
	for {
		select {
		case msg := <-lastMsgChannel:
			// triggered when a producer is done
			{
				if string(msg.Body) == "end" {
					nProducersDone++
					log.Println(nProducersDone, "producers are done")
					if nProducersDone == nProdTotal {
						log.Println("all producers are done, telling supervisor to stop workers")
						// tell supervisor to stop workers
						close(done)
						goto EndOfExperiment
					}
				} else {
					utils.Handle(errors.New(`received message with body != "end" through lastMsg exchange, something went wrong`))
				}
			}
		case <-interrupt:
			// triggered when interrupt comes from this machine
			// when an interrupt happens, the consumer should also stop
			// => but it doesn't need to tell the supervisor + workers anymore
			{
				log.Println("consumer interrupted, waiting for workers to finish before stopping")
				goto EndOfExperiment
			}
		}
	}

EndOfExperiment:
	log.Println("consumer reached end of experiment, waiting for workers to finish")
	// wait for all workers to actually finish
	wg.Wait()
	log.Println("consumer: wait group done")
}

// DEPRECATED
func (c *Consumer) Start(interrupt <-chan os.Signal) {

	// connect to the broker
	connection, err := amqp.Dial(c.config.Broker.URL)
	utils.Handle(err)
	defer connection.Close()

	// create a channel to access api
	channel, err := connection.Channel()
	utils.Handle(err)
	defer channel.Close()

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
	stop := make(chan bool)
	workers := make([]*BufWorker, c.config.Consumer.NWorkers)
	for i := 0; i < len(workers); i++ {
		/*workers[i] = NewWorker(
			connection,
			fmt.Sprintf("worker-%d", i),
			c.config,
			queue,
		)*/
		workers[i] = NewBufWorker(
			fmt.Sprintf("worker-%d", i),
			c.config,
			connection,
			queue,
		)
	}

	var wg sync.WaitGroup

	// start the workers
	for _, worker := range workers {
		//wId := worker.workerId
		wId := worker.id
		worker := worker
		go func() {
			wg.Add(1)
			log.Printf("worker \"%s\" started\n", wId)
			worker.Start(stop)
			wg.Done()
		}()
	}

	// each producer has to send "end" to each (i.e. also to this) consumer
	// count how many times that happens and stop workers when all producers are done
	// also: we could get an interrupt to which we need to pay attention
	nProducersDone, prodTotal := 0, c.config.Producer.NProducers
	for {
		select {
		case msg := <-lastMsgChannel:
			{
				if string(msg.Body) == "end" {
					nProducersDone++
					log.Printf("received \"end\" from producer, %d/%d producers are done\n", nProducersDone, prodTotal)
					if nProducersDone == prodTotal {
						log.Println("all producers are done, stopping ...")
						goto TheEnd
					}
				}
			}
		case <-interrupt:
			{
				log.Println("consumer was interrupted, stopping workers")
				goto TheEnd
			}
		}
	}
TheEnd:
	close(stop)

	// wait here until the workers are really done
	wg.Wait()
	log.Println("workers wait group done")
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
