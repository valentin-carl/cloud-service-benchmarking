package measurements

import (
	"benchmark/lib/utils"
	"errors"
	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	"log"
)

type Measurement interface {
	Record() []string
}

// NewMeasurement
//   - t: is the current timestamp, unix ms
//   - stats: is variadic because some measurements might need values for
//     multiple points in time to calculate utilization (e.g., CPU)
//     => otherwise, just use the first element
//     (function panics if there is no first element)
//
// todo extend with network stats
func NewMeasurement[T interface{ *cpu.Stats | *memory.Stats }](t int64, stats ...T) Measurement {

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
	default:
		{
			// this should never be reached
			panic("problem with measurement types")
		}
	}
}
