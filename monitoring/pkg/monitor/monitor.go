package monitor

import (
	"benchmark/lib/utils"
	"encoding/csv"
	"errors"
	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	"log"
	"monitoring/pkg/measurements"
	"os"
	"time"
)

type Monitor[T interface{ *cpu.Stats | *memory.Stats }] struct {
	w    *csv.Writer // to access the measurements file
	prev T           // previous measurement (might be interesting for some kinds of measurements, e.g., CPU)
}

func NewMonitor[T interface{ *cpu.Stats | *memory.Stats }](file *os.File) *Monitor[T] {
	return &Monitor[T]{
		w: csv.NewWriter(file),
	}
}

func (m *Monitor[T]) Start(timer <-chan time.Time, stop <-chan bool) {

	// make sure no data is lost when stopping
	defer func() {
		log.Println("cpu monitor executing deferred flush")
		m.w.Flush()
	}()

	// write csv column names
	var columnNames []string
	switch any(*new(T)).(type) { // it's so stupid that this works :-D
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
				}

				// create measurement
				log.Println("creating new cpu measurement")
				measurement := measurements.NewMeasurement[T](time.Now().UnixMilli(), curr, m.prev)

				// store measurement
				err = m.w.Write(measurement.Record())
				utils.Handle(err)

				// store old stats
				m.prev = curr
			}
		case <-stop:
			{
				log.Println("CPU monitor received stop signal")
				goto TheEnd
			}
		}
	}
TheEnd:
	log.Println("CPU monitor done, flushing remaining buffer contents")
}
