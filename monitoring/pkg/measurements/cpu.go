package measurements

import (
	"fmt"
	"github.com/mackerelio/go-osstat/cpu"
	"log"
	"math"
)

var (
	ColumnNamesCPU = []string{
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

type CPUMeasurement struct {
	timestamp                       int64   // unix timestamp of measurement
	user, system, idle, nice, total uint64  // raw values
	userp, systemp, idlep           float64 // percentage calculated with last measurement
}

func NewCPUMeasurement(
	timestamp int64, // todo for analysis: these values are milliseconds => convert them to seconds
	curr *cpu.Stats,
	prev *cpu.Stats,
) Measurement {

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

	return &CPUMeasurement{
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

func (m *CPUMeasurement) Record() []string {
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
