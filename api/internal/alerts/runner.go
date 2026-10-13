package alerts

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"time"
)

type Queue interface {
	DeliverOne(context.Context, Send) (bool, error)
}

func Run(ctx context.Context, q Queue, send Send) {
	for ctx.Err() == nil {
		op, cancel := context.WithTimeout(ctx, 15*time.Second)
		worked, err := q.DeliverOne(op, send)
		cancel()
		delay := 250 * time.Millisecond
		if !worked || err != nil {
			delay = 5 * time.Second
		}
		if err != nil && ctx.Err() == nil {
			slog.Error("entrega: persistência não concluída")
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}
func RetryDelay(attempt int) time.Duration {
	intervals := []time.Duration{time.Minute, 5 * time.Minute, 15 * time.Minute, time.Hour}
	if attempt < 1 {
		attempt = 1
	}
	if attempt > 4 {
		attempt = 4
	}
	base := intervals[attempt-1]
	return base + time.Duration(rand.Int64N(int64(base/10)+1))
}
