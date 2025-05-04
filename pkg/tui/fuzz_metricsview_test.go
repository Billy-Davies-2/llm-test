package tui

import (
	"testing"
)

// minimal model stub so we can call metricsView
type fakeModel struct {
	metricsMap    map[string]MetricsData
	width, height int
}

func (m fakeModel) metricsView() string {
	// copy or alias your real implementation here
	// e.g. return realModel(m).metricsView()
	return "" // replace with actual call
}

func FuzzMetricsView(f *testing.F) {
	// seed with a normal small map
	f.Add(map[string]MetricsData{
		"a": {HostID: "a", CPU: 0.5, RAM: 0.7},
		"b": {HostID: "b", CPU: 0.1, RAM: 0.2},
	}, 80, 24)

	f.Fuzz(func(t *testing.T, mMap map[string]MetricsData, w, h int) {
		// guard against zero or negative sizes
		if w < 1 || h < 1 {
			return
		}
		fm := fakeModel{metricsMap: mMap, width: w, height: h}
		_ = fm.metricsView() // must not panic
	})
}
