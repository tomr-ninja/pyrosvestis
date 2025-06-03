package pyrosvestis

import (
	"fmt"
	"net/http"
)

func FetchMetrics(endpoint string) (*Snapshot, error) {
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch metrics: %s", resp.Status)
	}

	snapshot := NewSnapshot()
	if err = snapshot.Decode(resp.Body); err != nil {
		return nil, fmt.Errorf("failed to decode metrics: %w", err)
	}

	return snapshot, nil
}
