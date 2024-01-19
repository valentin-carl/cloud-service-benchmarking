package measurements

import (
	"fmt"
	"github.com/mackerelio/go-osstat/memory"
)

var (
	ColumnNamesMemory = []string{
		"timestamp",
		"free",
		"total",
		"active",
		"cached",
		"inactive",
		"swapFree",
		"swapTotal",
		"swapUsed",
		"used",
		"freep",
	}
)

type MemoryMeasurement struct {
	timestamp                                                                  int64   // unix timestamp of measurement
	free, total, active, cached, inactive, swapFree, swapTotal, swapUsed, used uint64  // values in bytes
	freep                                                                      float64 // freep => free/total * 100
}

func NewMemoryMeasurement(timestamp int64, stats *memory.Stats) Measurement {
	return &MemoryMeasurement{
		timestamp: timestamp,
		free:      stats.Free,
		total:     stats.Total,
		active:    stats.Active,
		cached:    stats.Cached,
		inactive:  stats.Inactive,
		swapFree:  stats.SwapFree,
		swapTotal: stats.SwapTotal,
		swapUsed:  stats.SwapUsed,
		used:      stats.Used,
		freep:     float64(stats.Free) / float64(stats.Total) * 100,
	}
}

func (m *MemoryMeasurement) Records() [][]string {
	return [][]string{{
		fmt.Sprintf("%d", m.timestamp),
		fmt.Sprintf("%d", m.free),
		fmt.Sprintf("%d", m.total),
		fmt.Sprintf("%d", m.active),
		fmt.Sprintf("%d", m.cached),
		fmt.Sprintf("%d", m.inactive),
		fmt.Sprintf("%d", m.swapFree),
		fmt.Sprintf("%d", m.swapTotal),
		fmt.Sprintf("%d", m.swapUsed),
		fmt.Sprintf("%d", m.used),
		fmt.Sprintf("%f", m.freep),
	}}
}
