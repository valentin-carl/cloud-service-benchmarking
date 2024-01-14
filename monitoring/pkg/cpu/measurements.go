package cpu

import (
	"benchmark/lib/utils"
	"encoding/csv"
	"fmt"
	"github.com/mackerelio/go-osstat/cpu"
	"log"
	"math"
	"os"
	"time"
)

var (
	columnNames = []string{
		"timestamp",
		"user",
		"system",
		"idle",
		"nice",
		"total",
		"userp",
		"systemp",
		"idlep",
	}
)

type Monitor struct {
	w *csv.Writer
}

// todo create the csv somewhere else
func NewMonitor(file *os.File) *Monitor {
	return &Monitor{
		w: csv.NewWriter(file),
	}
}

// Start should be used as `go measurement.Start(...)`
func (m *Monitor) Start(timer <-chan time.Time, stop <-chan bool) {

	defer m.w.Flush()

	// write csv column names
	err := m.w.Write(columnNames)
	utils.Handle(err)
	m.w.Flush()

	var prev *cpu.Stats

	for {
		select {
		case <-timer:
			{
				// get + store new measurements
				stats, err := cpu.Get()
				utils.Handle(err)
				log.Println("storing new measurement") // todo comment out after debugging
				currentMeasurement := newMeasurement(time.Now().UnixMilli(), stats, prev)
				err = m.w.Write(currentMeasurement.record())
				utils.Handle(err)
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

//
// This measurement type is not exported because it is only used by
// the CPU monitor, other monitors will have their own internal
// measurement representation.
//

type measurement struct {
	timestamp                       int64   // unix timestamp of measurement
	user, system, idle, nice, total uint64  // raw values
	userp, systemp, idlep           float64 // percentage calculated with last measurement
}

func newMeasurement(
	timestamp int64, // todo for analysis: these values are millisecond => convert them to seconds
	curr *cpu.Stats,
	prev *cpu.Stats,
) *measurement {

	var userp, systemp, idlep float64

	if prev == nil {
		log.Println("no previous measurement, unable to calculate relative values")
		userp, systemp, idlep = math.NaN(), math.NaN(), math.NaN()
	} else {
		log.Println("previous measurement available, calculating relative values")
		tDiff := float64(curr.Total - prev.Total)
		userp = (float64(curr.User-prev.User) / tDiff) * 100
		systemp = (float64(curr.System-prev.System) / tDiff) * 100
		idlep = (float64(curr.Idle-prev.Idle) / tDiff) * 100
	}

	return &measurement{
		timestamp: timestamp,
		user:      curr.User,
		system:    curr.System,
		idle:      curr.Idle,
		nice:      curr.Nice,
		total:     curr.Total,
		userp:     userp,
		systemp:   systemp,
		idlep:     idlep,
	}
}

func (m *measurement) record() []string {
	return []string{
		fmt.Sprintf("%d", m.timestamp),
		fmt.Sprintf("%d", m.user),
		fmt.Sprintf("%d", m.system),
		fmt.Sprintf("%d", m.idle),
		fmt.Sprintf("%d", m.nice),
		fmt.Sprintf("%d", m.total),
		fmt.Sprintf("%f", m.userp),
		fmt.Sprintf("%f", m.systemp),
		fmt.Sprintf("%f", m.idlep),
	}
}
