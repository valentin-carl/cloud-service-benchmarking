package main

import (
	"encoding/csv"
	"fmt"
	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	"log"
	"math"
	"os"
	"time"
)

const (
	dir      = "./data"
	filename = "measurements-0.csv"
)

func Handle(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

	{
		// get cpu stats
		x, err := cpu.Get()
		if err != nil {
			panic(err)
		}
		fmt.Println(x)
		fmt.Println(x.Total)
		fmt.Println(x.Nice)
		fmt.Println(x.Idle)
		fmt.Println(x.System)
		fmt.Println(x.User)

		// get memory stats
		y, err := memory.Get()
		if err != nil {
			panic(err)
		}
		fmt.Println(y.Free)
	}

	//
	//
	//

	err := os.Mkdir("./data", os.ModePerm)
	if err != nil {
		log.Println("dir", dir, "already exists")
	}

	file, err := os.Create(fmt.Sprintf("%s/%s", dir, filename)) // todo create new file instead of truncating the old one
	Handle(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	// write csv column names
	err = w.Write([]string{
		"timestamp",
		"user",
		"system",
		"idle",
		"nice",
		"total",
		"userp",
		"systemp",
		"idlep",
	})
	Handle(err)
	w.Flush()

	var prev *cpu.Stats
	t := time.NewTicker(time.Second)

	stop := time.NewTimer(15 * time.Second)

	for {
		select {
		case <-t.C:
			{
				// get + store new measurements
				curr, err := cpu.Get()
				Handle(err)
				m := NewMeasurement(time.Now().UnixMilli(), curr, prev)
				err = m.Write(w)
				Handle(err)
				prev = curr
			}
		case <-stop.C:
			{
				log.Println("received stop signal")
				goto TheEnd
			}
		}
	}

TheEnd:
	log.Println("done")
}

type Measurement struct {
	timestamp                       int64   // unix timestamp of measurement
	user, system, idle, nice, total uint64  // raw values
	userp, systemp, idlep           float64 // percentage calculated with last measurement
}

func NewMeasurement(
	timestamp int64, // todo for analysis: these values are millisecond => convert them to seconds
	curr *cpu.Stats,
	prev *cpu.Stats,
) *Measurement {

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

	return &Measurement{
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

func (m *Measurement) Write(w *csv.Writer) error {
	return w.Write(m.Record())
}

func (m *Measurement) Record() []string {
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
