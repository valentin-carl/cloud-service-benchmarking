package measurements

import (
	"benchmark/lib/utils"
	"errors"
	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/mackerelio/go-osstat/network"
	"log"
)

type Measurement interface {
	Records() [][]string // one record == one line in csv (for some kinds of measurements, there might be multiple record per timestamp)
}

// NewMeasurement
//   - t: is the current timestamp, unix ms // todo for analysis: convert ms to s
//   - stats: is variadic because some measurements might need values for
//     multiple points in time to calculate utilization (e.g., CPU)
//     => otherwise, just use the first element
//     (function panics if there is no first element)
func NewMeasurement[T interface {
	*cpu.Stats | *memory.Stats | []network.Stats
}](t int64, stats ...T) Measurement {

	// quick safety check
	if len(stats) == 0 {
		utils.Handle(errors.New("no stats to create measurement"))
	}

	// create the measurement object
	switch any(stats[0]).(type) {
	case *cpu.Stats:
		{
			log.Println("measurement.go: creating new cpu measurement")
			s0, ok := any(stats[0]).(*cpu.Stats)
			if !ok {
				utils.Handle(errors.New("could not convert cpu measurement to *cpu.Stats"))
			}
			// note: stats[1] might be nil (but len(stats) == 2 is true nonetheless!)
			s1, ok := any(stats[1]).(*cpu.Stats)
			if !ok {
				utils.Handle(errors.New("could not convert cpu measurement to *cpu.Stats"))
			}
			return NewCPUMeasurement(t, s0, s1)
		}
	case *memory.Stats:
		{
			log.Println("creating new memory measurement")
			s, ok := any(stats[0]).(*memory.Stats)
			if !ok {
				utils.Handle(errors.New("could not convert cpu measurement to *cpu.Stats"))
			}
			return NewMemoryMeasurement(t, s)
		}
	case []network.Stats:
		{
			log.Println("creating new memory measurement")
			s0, ok := any(stats[0]).([]network.Stats)
			if !ok {
				utils.Handle(errors.New("could not convert cpu measurement to *network.Stats"))
			}
			s1, ok := any(stats[1]).([]network.Stats)
			if !ok {
				utils.Handle(errors.New("could not convert cpu measurement to *network.Stats"))
			}
			return NewNetworkMeasurement(t, s0, s1)
		}
	default:
		{
			// this should never be reached
			panic("problem with measurement types")
		}
	}
}
