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
		"swapUsed",
		"used",
	}
)

type MemoryMeasurement struct {
	timestamp                                                                  int64  // unix timestamp of measurement
	free, total, active, cached, inactive, swapFree, swapTotal, swapUsed, used uint64 // todo should be in bytes but check again
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
		swapUsed:  stats.SwapUsed,
		used:      stats.Used,
	}
}

func (m *MemoryMeasurement) Record() []string {
	return []string{
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
	}
}
