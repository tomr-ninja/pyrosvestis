package pyrosvestis

import (
	"errors"
	"io"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

type Snapshot struct {
	Metrics      []MetricsGroup // Slice of metrics in the snapshot
	reverseIndex map[string]int // Maps metric names to their index in the snapshot
}

type MetricsGroup struct {
	Name    string              // Name of the metric group
	Metrics []*dto.MetricFamily // Slice of metrics in the group (different labels)
}

// NewSnapshot creates a new Snapshot instance.
func NewSnapshot() *Snapshot {
	return &Snapshot{
		Metrics:      make([]MetricsGroup, 0),
		reverseIndex: make(map[string]int),
	}
}

// Decode reads metrics from the provided io.Reader and populates the Snapshot.
func (s *Snapshot) Decode(r io.Reader) error {
	s.Metrics = make([]MetricsGroup, 0)
	s.reverseIndex = make(map[string]int)

	dec := expfmt.NewDecoder(r, expfmt.NewFormat(expfmt.TypeTextPlain))
	for {
		var metric dto.MetricFamily
		if err := dec.Decode(&metric); err != nil {
			if errors.Is(err, io.EOF) {
				break // End of the stream
			}

			return err // Return any other error encountered during decoding
		}

		name := metric.GetName()
		if index, exists := s.reverseIndex[name]; !exists {
			index = len(s.Metrics)
			s.reverseIndex[name] = index
			s.Metrics = append(s.Metrics, MetricsGroup{Name: name, Metrics: []*dto.MetricFamily{&metric}})
		} else {
			s.Metrics[index].Metrics = append(s.Metrics[index].Metrics, &metric)
		}
	}

	return nil
}
