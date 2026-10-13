package telemetry

import (
	"strings"
	"testing"
)

func TestPrometheusFormat(t *testing.T) {
	text := Render(Snapshot{Checks: 2, DurationMS: 1500, Buckets: []int64{0, 1, 2, 2, 2, 2, 2, 2}, Endpoints: 1, Healthy: 1})
	for _, v := range []string{"# TYPE certguard_check_duration_seconds histogram\n", "certguard_check_duration_seconds_sum 1.5\n", "certguard_check_duration_seconds_bucket{le=\"+Inf\"} 2\n", "certguard_endpoints_by_status{status=\"healthy\"} 1\n"} {
		if !strings.Contains(text, v) {
			t.Fatal(v, text)
		}
	}
	for _, forbidden := range []string{"hostname=", "email=", "endpoint_id=", "channel_id=", "url=", "token=", "code="} {
		if strings.Contains(text, forbidden) {
			t.Fatal("etiqueta sensível", forbidden)
		}
	}
	lines := strings.Split(strings.TrimSpace(text), "\n")
	series := 0
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") {
			series++
		}
	}
	if series != 30 {
		t.Fatal("cardinalidade alterada", series)
	}
}
