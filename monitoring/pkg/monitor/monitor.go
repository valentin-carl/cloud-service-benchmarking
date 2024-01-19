package monitor

import (
	"benchmark/lib/utils"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/mackerelio/go-osstat/network"
	"log"
	"monitoring/pkg/measurements"
	"os"
	"time"
)

type Monitor[T interface {
	*cpu.Stats | *memory.Stats | []network.Stats
}] struct {
	w    *csv.Writer // to access the measurements file
	prev T           // previous measurement (might be interesting for some kinds of measurements, e.g., CPU)
	name string      // just for logging
}

func NewMonitor[T interface {
	*cpu.Stats | *memory.Stats | []network.Stats
}](file *os.File) *Monitor[T] {
	return &Monitor[T]{
		w:    csv.NewWriter(file),
		name: fmt.Sprintf("%T-monitor", *new(T)),
	}
}

func (m *Monitor[T]) Start(timer <-chan time.Time, stop <-chan bool) {

	// make sure no data is lost when stopping
	defer func() {
		log.Println(m.name, "executing deferred flush")
		m.w.Flush()
	}()

	// write csv column names
	var columnNames []string
	switch any(*new(T)).(type) {
	case *cpu.Stats:
		{
			log.Println("writing CPU column names")
			columnNames = measurements.ColumnNamesCPU
		}
	case *memory.Stats:
		{
			log.Println("writing memory column names")
			columnNames = measurements.ColumnNamesMemory
		}
	case []network.Stats:
		{
			log.Println("writing network column names")
			columnNames = measurements.ColumnNamesNetwork
		}
	default:
		{
			// this shouldn't be reached
			utils.Handle(errors.New("invalid measurement type"))
		}
	}
	err := m.w.Write(columnNames)
	utils.Handle(err)
	m.w.Flush()

	// continuously get + store measurements
	for {
		select {
		case <-timer:
			{
				var curr T
				switch any(*new(T)).(type) {
				case *cpu.Stats:
					{
						// get cpu stats
						cpuStats, err := cpu.Get()
						utils.Handle(err)
						curr = any(cpuStats).(T)
						log.Println("got cpu stats")
					}
				case *memory.Stats:
					{
						// get memory stats
						memStats, err := memory.Get()
						utils.Handle(err)
						curr = any(memStats).(T)
						log.Println("got memory stats")
					}
				case []network.Stats:
					{
						// get network stats
						netStats, err := network.Get()
						utils.Handle(err)
						curr = any(netStats).(T)
						log.Println("got network stats")
					}
				default:
					{
						// this shouldn't be reached
						utils.Handle(errors.New("invalid measurement type"))
					}
				}

				// create measurement
				log.Println(m.name, "creating new measurement")
				measurement := measurements.NewMeasurement[T](time.Now().UnixMilli(), curr, m.prev)

				// store measurement(s)
				if m.prev != nil {
					for _, record := range measurement.Records() {
						err = m.w.Write(record)
						utils.Handle(err)
					}
				} else {
					log.Println("no previous measurement, skipping write ...")
				}

				// store old stats
				m.prev = curr
			}
		case <-stop:
			{
				log.Println(m.name, "received stop signal")
				goto TheEnd
			}
		}
	}
TheEnd:
	log.Println(m.name, "done, flushing remaining buffer contents")
}
