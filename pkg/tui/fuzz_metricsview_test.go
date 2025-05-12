package tui

import (
	"encoding/json"
	"testing"
)

type MetricsData struct {
	HostID string
	CPU    float64
	RAM    float64
}

type fakeModel struct {
	metricsMap    map[string]MetricsData
	width, height int
}

func (m fakeModel) metricsView() string {
	// Copy or alias your real implementation here
	// e.g., return realModel(m).metricsView()
	return "" // Replace with actual call
}

func FuzzMetricsView(f *testing.F) {
	// Seed with a normal small map
	sampleMap := map[string]MetricsData{
		"a": {HostID: "a", CPU: 0.5, RAM: 0.7},
		"b": {HostID: "b", CPU: 0.1, RAM: 0.2},
	}
	sampleJSON, err := json.Marshal(sampleMap)
	if err != nil {
		f.Fatal(err) // Fail the test if seeding fails
	}
	f.Add(sampleJSON, 80, 24) // Seed with JSON []byte, width, and height

	f.Fuzz(func(t *testing.T, mapJSON []byte, w, h int) {
		// Guard against zero or negative sizes
		if w < 1 || h < 1 {
			return
		}
		// Attempt to unmarshal the fuzzed []byte into a map
		var mMap map[string]MetricsData
		if err := json.Unmarshal(mapJSON, &mMap); err != nil {
			return // Skip invalid JSON inputs
		}
		// Optional: Skip large maps to avoid performance issues
		if len(mMap) > 100 {
			return
		}
		// Construct the model and test metricsView
		fm := fakeModel{metricsMap: mMap, width: w, height: h}
		_ = fm.metricsView() // Must not panic
	})
}
