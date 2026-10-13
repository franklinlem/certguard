package httpapi

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"github.com/franklinlem/certguard/api/internal/telemetry"
	"net/http"
	"strings"
	"time"
)

func ValidateMetricsToken(token string) error {
	if token == "" {
		return nil
	}
	if len(token) < 32 || len(token) > 256 {
		return errors.New("METRICS_TOKEN deve conter 32 a 256 caracteres ASCII sem espaços")
	}
	for _, r := range token {
		if r < 33 || r > 126 {
			return errors.New("METRICS_TOKEN inválido")
		}
	}
	return nil
}
func MetricsHandler(token string, collect func(context.Context) (telemetry.Snapshot, error)) http.Handler {
	expected := sha256.Sum256([]byte(token))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if token == "" || ValidateMetricsToken(token) != nil {
			http.NotFound(w, r)
			return
		}
		header := r.Header.Values("Authorization")
		provided := ""
		if len(header) == 1 {
			parts := strings.SplitN(header[0], " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && len(parts[1]) <= 256 {
				provided = parts[1]
			}
		}
		digest := sha256.Sum256([]byte(provided))
		if subtle.ConstantTimeCompare(expected[:], digest[:]) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			fail(w, 401, "Acesso às métricas não autorizado.")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		snapshot, err := collect(ctx)
		if err != nil {
			fail(w, 503, "Métricas temporariamente indisponíveis.")
			return
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = w.Write([]byte(telemetry.Render(snapshot)))
	})
}
