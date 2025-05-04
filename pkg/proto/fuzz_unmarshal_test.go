package proto

import (
	"testing"

	// classic protobuf API
	"google.golang.org/protobuf/proto"
)

// seed corpus with one valid MetricsResponse.
func FuzzUnmarshalMetrics(f *testing.F) {
	orig := &MetricsResponse{
		HostId:          "seed-host",
		CpuUsagePercent: 12.34,
		MemoryUsedMb:    256,
		MemoryTotalMb:   512,
		Gpu: &GPUInfo{
			Name:               "SeedGPU",
			TemperatureCelsius: 55.5,
		},
	}
	data, _ := proto.Marshal(orig)
	f.Add(data)

	f.Fuzz(func(t *testing.T, in []byte) {
		var m MetricsResponse
		// Should never panic or allocate unboundedly
		_ = proto.Unmarshal(in, &m)
	})
}
