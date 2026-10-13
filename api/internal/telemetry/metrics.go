// Package telemetry formata métricas operacionais sem identificadores sensíveis.
package telemetry

import (
	"fmt"
	"strings"
)

type Snapshot struct {
	Checks               int64
	CheckFailures        int64
	DurationMS           int64
	Buckets              []int64
	Deliveries           int64
	DeliveryFailures     int64
	DeliveryAccepted     int64
	Endpoints            int64
	Enabled              int64
	Paused               int64
	Healthy              int64
	Warning              int64
	Critical             int64
	Errors               int64
	Unchecked            int64
	Expiring30           int64
	Due                  int64
	PendingDeliveries    int64
	FailedDeliveries     int64
	OldestPendingSeconds float64
	LastCheckUnix        float64
	LastDeliveryUnix     float64
}

func Render(m Snapshot) string {
	var b strings.Builder
	counter := func(name, help string, v int64) {
		fmt.Fprintf(&b, "# HELP %s %s\n# TYPE %s counter\n%s %d\n", name, help, name, name, v)
	}
	gauge := func(name, help string, v float64) {
		fmt.Fprintf(&b, "# HELP %s %s\n# TYPE %s gauge\n%s %g\n", name, help, name, name, v)
	}
	counter("certguard_checks_total", "Verificacoes persistidas desde a inicializacao dos contadores.", m.Checks)
	counter("certguard_check_failures_total", "Verificacoes com code diferente de success.", m.CheckFailures)
	counter("certguard_alert_deliveries_total", "Tentativas de entrega persistidas.", m.Deliveries)
	counter("certguard_alert_delivery_failures_total", "Tentativas sem aceite remoto.", m.DeliveryFailures)
	counter("certguard_alert_delivery_accepted_total", "Entregas com aceite remoto persistido.", m.DeliveryAccepted)
	fmt.Fprintln(&b, "# HELP certguard_check_duration_seconds Duracao das verificacoes persistidas em segundos.\n# TYPE certguard_check_duration_seconds histogram")
	for i, bound := range []string{"0.1", "0.5", "1", "2.5", "5", "10", "30", "+Inf"} {
		var count int64
		if i < len(m.Buckets) {
			count = m.Buckets[i]
		}
		fmt.Fprintf(&b, "certguard_check_duration_seconds_bucket{le=%q} %d\n", bound, count)
	}
	fmt.Fprintf(&b, "certguard_check_duration_seconds_sum %g\ncertguard_check_duration_seconds_count %d\n", float64(m.DurationMS)/1000, m.Checks)
	gauge("certguard_endpoints", "Endpoints nao arquivados, incluindo pausados.", float64(m.Endpoints))
	fmt.Fprintln(&b, "# HELP certguard_endpoints_by_status Ultimo resultado do destino atual.\n# TYPE certguard_endpoints_by_status gauge")
	for _, v := range []struct {
		name  string
		value int64
	}{{"healthy", m.Healthy}, {"warning", m.Warning}, {"critical", m.Critical}, {"error", m.Errors}, {"unchecked", m.Unchecked}} {
		fmt.Fprintf(&b, "certguard_endpoints_by_status{status=%q} %d\n", v.name, v.value)
	}
	gauge("certguard_endpoints_enabled", "Endpoints ativos nao arquivados.", float64(m.Enabled))
	gauge("certguard_endpoints_paused", "Endpoints pausados nao arquivados.", float64(m.Paused))
	gauge("certguard_certificates_expiring", "Ultimo certificado observado vence entre agora e 30 dias.", float64(m.Expiring30))
	gauge("certguard_endpoints_due", "Endpoints ativos com agenda vencida.", float64(m.Due))
	gauge("certguard_alert_deliveries_pending", "Entregas pendentes, inclusive com executor desligado.", float64(m.PendingDeliveries))
	gauge("certguard_alert_deliveries_failed", "Entregas em falha definitiva.", float64(m.FailedDeliveries))
	gauge("certguard_alert_oldest_pending_age_seconds", "Idade da pendencia mais antiga, zero se nenhuma.", m.OldestPendingSeconds)
	gauge("certguard_last_check_timestamp_seconds", "Ultimo check persistido, zero se nenhum.", m.LastCheckUnix)
	gauge("certguard_last_delivery_attempt_timestamp_seconds", "Ultima tentativa persistida, zero se nenhuma.", m.LastDeliveryUnix)
	return b.String()
}
