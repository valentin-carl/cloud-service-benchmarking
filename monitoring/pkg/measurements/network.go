package measurements

import (
	"fmt"
	"github.com/mackerelio/go-osstat/network"
	"log"
)

var (
	ColumnNamesNetwork = []string{
		"timestamp",
		"name",
		"RxBytes",
		"TxBytes",
	}
)

type NetworkMeasurement struct {
	timestamp    int64
	measurements map[string]struct {
		RxBytes, TxBytes uint64 // bytes received/transmitted since the previous measurement
	}
}

func NewNetworkMeasurement(timestamp int64, current, previous []network.Stats) Measurement {

	// small helper to make my life easier :-)
	toMap := func(measurement []network.Stats) map[string]struct{ RxBytes, TxBytes uint64 } {
		res := make(map[string]struct{ RxBytes, TxBytes uint64 })
		for _, s := range measurement {
			res[s.Name] = struct{ RxBytes, TxBytes uint64 }{RxBytes: s.RxBytes, TxBytes: s.TxBytes}
		}
		return res
	}

	// calculate traffic between previous and current measurements
	// only collect non-zero values
	curr, prev := toMap(current), toMap(previous)
	m := make(map[string]struct{ RxBytes, TxBytes uint64 })
	for name, vals := range curr {
		if pVals, ok := prev[name]; ok {
			log.Println("found", name, "in previous measurement")
			rxDiff, txDiff := vals.RxBytes-pVals.RxBytes, vals.TxBytes-pVals.TxBytes
			log.Printf("bytes received: %d, bytes transmitted: %d\n", rxDiff, txDiff)
			if rxDiff != 0 || txDiff != 0 {
				m[name] = struct{ RxBytes, TxBytes uint64 }{RxBytes: rxDiff, TxBytes: txDiff}
			}
		} else {
			log.Println(name, "not in prev network stats, storing absolute values")
			// new network interface found => store absolute values because it would have been zero for the last measurement
			// todo should be fine but think about this again
			m[name] = struct{ RxBytes, TxBytes uint64 }{RxBytes: vals.RxBytes, TxBytes: vals.TxBytes}
		}
	}

	return &NetworkMeasurement{
		timestamp:    timestamp,
		measurements: m,
	}
}

func (m *NetworkMeasurement) Records() [][]string {
	res := make([][]string, len(m.measurements))
	i := 0
	for name, values := range m.measurements {
		res[i] = []string{
			fmt.Sprintf("%d", m.timestamp),
			name,
			fmt.Sprintf("%d", values.RxBytes),
			fmt.Sprintf("%d", values.TxBytes),
		}
		i++
	}
	return res
}
