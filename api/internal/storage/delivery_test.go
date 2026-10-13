package storage

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/franklinlem/certguard/api/internal/alerts"
	"github.com/franklinlem/certguard/api/internal/certificate"
	"github.com/franklinlem/certguard/api/internal/testutil"
	"strings"
	"testing"
	"time"
)

func deliveryStore(t *testing.T) *Store {
	t.Helper()
	c, err := alerts.NewCipher(base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{9}, 32)))
	if err != nil {
		t.Fatal(err)
	}
	return &Store{Pool: testutil.Database(t), Cipher: c}
}
func channelInput() ChannelInput {
	return ChannelInput{Name: "Teste", Type: "webhook", Config: alerts.ChannelConfig{URL: "https://notify.example.com/hook", Secret: strings.Repeat("s", 32)}}
}
func activeChannel(t *testing.T, s *Store) Channel {
	t.Helper()
	c, err := s.CreateChannel(context.Background(), channelInput())
	if err != nil {
		t.Fatal(err)
	}
	yes := true
	c, err = s.UpdateChannel(context.Background(), c.ID, nil, &yes, false)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func TestPostgresDeliveryLifecycle(t *testing.T) {
	s := deliveryStore(t)
	ctx := context.Background()
	e := newEndpoint(t, s, "example.com")
	c, err := s.CreateChannel(ctx, channelInput())
	if err != nil || c.Enabled {
		t.Fatal(c, err)
	}
	checkAlert(t, s, e, "warning", 20, "a")
	list, err := s.Deliveries(ctx, "", 20, 0)
	if err != nil || len(list) != 0 {
		t.Fatal("canal pausado recebeu evento", list, err)
	}
	yes := true
	c, err = s.UpdateChannel(ctx, c.ID, nil, &yes, false)
	if err != nil {
		t.Fatal(err)
	}
	checkAlert(t, s, e, "critical", 5, "a")
	checkAlert(t, s, e, "critical", 4, "a")
	list, err = s.Deliveries(ctx, "", 20, 0)
	if err != nil || len(list) != 1 || list[0].Attempts != 0 {
		t.Fatal(list, err)
	}
	id := list[0].ID
	var sealed []byte
	if err = s.Pool.QueryRow(ctx, "SELECT config_encrypted FROM alert_deliveries WHERE id=$1", id).Scan(&sealed); err != nil || bytes.Contains(sealed, []byte(channelInput().Config.Secret)) {
		t.Fatal("segredo exposto", err)
	}
	calls := 0
	send := func(_ context.Context, kind string, cfg alerts.ChannelConfig, body []byte, delivery string) alerts.Outcome {
		calls++
		if kind != "webhook" || cfg.Secret != channelInput().Config.Secret || delivery != id {
			t.Error("snapshot incorreto")
		}
		var p alerts.Payload
		if json.Unmarshal(body, &p) != nil || p.Event != "certificate_expiry_critical" {
			t.Error("payload incorreto")
		}
		return alerts.Outcome{Retry: true, Code: "http_503"}
	}
	if worked, err := s.DeliverOne(ctx, send); err != nil || !worked {
		t.Fatal(worked, err)
	}
	list, err = s.Deliveries(ctx, "", 20, 0)
	if err != nil || list[0].Attempts != 1 || list[0].Status != "pending" || time.Until(*list[0].NextAttempt) < 50*time.Second {
		t.Fatal(list, err)
	}
	if worked, err := s.DeliverOne(ctx, send); err != nil || worked || calls != 1 {
		t.Fatal("backoff ignorado", worked, err)
	}
	if _, err = s.Pool.Exec(ctx, "UPDATE alert_deliveries SET next_attempt_at=now()-interval '1 second' WHERE id=$1", id); err != nil {
		t.Fatal(err)
	}
	if worked, err := s.DeliverOne(ctx, func(context.Context, string, alerts.ChannelConfig, []byte, string) alerts.Outcome {
		return alerts.Outcome{Accepted: true}
	}); err != nil || !worked {
		t.Fatal(worked, err)
	}
	list, err = s.Deliveries(ctx, "", 20, 0)
	if err != nil || list[0].Attempts != 2 || list[0].Status != "sent" || list[0].NextAttempt != nil || list[0].AcceptedAt == nil || list[0].LastError != nil {
		t.Fatal(list, err)
	}
	checkAlert(t, s, e, "healthy", 90, "b") // Recuperação também gera entrega.
	no := false
	if _, err = s.UpdateChannel(ctx, c.ID, nil, &no, false); err != nil {
		t.Fatal(err)
	}
	list, err = s.Deliveries(ctx, "", 20, 0)
	if err != nil || len(list) != 2 || list[0].Status != "cancelled" {
		t.Fatal(list, err)
	}
	if _, err = s.UpdateChannel(ctx, c.ID, nil, &yes, false); err != nil {
		t.Fatal(err)
	}
	if worked, err := s.DeliverOne(ctx, send); err != nil || worked {
		t.Fatal("reativação reenviou cancelada", worked, err)
	}
}
func TestPostgresDeliveryLocksCancelAndLimits(t *testing.T) {
	s := deliveryStore(t)
	ctx := context.Background()
	c := activeChannel(t, s)
	e := newEndpoint(t, s, "example.com")
	checkAlert(t, s, e, "warning", 20, "a")
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		_, err := s.DeliverOne(ctx, func(context.Context, string, alerts.ChannelConfig, []byte, string) alerts.Outcome {
			close(started)
			<-release
			return alerts.Outcome{Code: "network_error", Retry: true}
		})
		done <- err
	}()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("envio não iniciou")
	}
	if worked, err := s.DeliverOne(ctx, func(context.Context, string, alerts.ChannelConfig, []byte, string) alerts.Outcome {
		t.Error("entrega duplicada")
		return alerts.Outcome{Accepted: true}
	}); err != nil || worked {
		t.Error(worked, err)
	}
	no := false
	if _, err := s.UpdateChannel(ctx, c.ID, nil, &no, false); !Conflict(err) {
		t.Error("canal em envio foi alterado", err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if _, err := s.Pool.Exec(ctx, "UPDATE alert_deliveries SET next_attempt_at=now()-interval '1 second'"); err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	if _, err := s.DeliverOne(cancelled, func(context.Context, string, alerts.ChannelConfig, []byte, string) alerts.Outcome {
		cancel()
		return alerts.Outcome{Accepted: true}
	}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	list, err := s.Deliveries(ctx, "", 20, 0)
	if err != nil || list[0].Attempts != 1 || list[0].Status != "pending" {
		t.Fatal("cancelamento persistiu tentativa", list, err)
	}
	if _, err = s.Pool.Exec(ctx, "UPDATE alert_deliveries SET attempt_count=4"); err != nil {
		t.Fatal(err)
	}
	if worked, err := s.DeliverOne(ctx, func(context.Context, string, alerts.ChannelConfig, []byte, string) alerts.Outcome {
		return alerts.Outcome{Retry: true, Code: "network_error"}
	}); err != nil || !worked {
		t.Fatal(worked, err)
	}
	list, err = s.Deliveries(ctx, "", 20, 0)
	if err != nil || list[0].Attempts != 5 || list[0].Status != "failed" {
		t.Fatal("tentativas ilimitadas", list, err)
	}
	if _, err = s.UpdateChannel(ctx, c.ID, nil, nil, true); err != nil {
		t.Fatal(err)
	}
	channels, err := s.Channels(ctx)
	if err != nil || len(channels) != 0 {
		t.Fatal(channels, err)
	}
	for range 20 {
		if _, err = s.CreateChannel(ctx, channelInput()); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = s.CreateChannel(ctx, channelInput()); !errors.Is(err, ErrBusy) {
		t.Fatal("limite de canais ignorado", err)
	}
}
func TestPostgresDeliveryAtomicityAndKey(t *testing.T) {
	s := deliveryStore(t)
	ctx := context.Background()
	c := activeChannel(t, s)
	e := newEndpoint(t, s, "example.com")
	if _, err := s.Pool.Exec(ctx, "ALTER TABLE alert_deliveries ADD CONSTRAINT reject_test CHECK(false)"); err != nil {
		t.Fatal(err)
	}
	due(t, s, e.ID)
	if _, err := s.Check(ctx, e.ID, func(context.Context, Endpoint) certificate.Result { return resultFor(e, "warning", 20, "a") }); err == nil {
		t.Fatal("fila rejeitada não reverteu check")
	}
	page, err := s.Alerts(ctx, "all", e.ID, 20, 0)
	if err != nil || len(page.Items) != 0 {
		t.Fatal(page, err)
	}
	history, err := s.History(ctx, e.ID, 20, 0)
	if err != nil || len(history) != 0 {
		t.Fatal(history, err)
	}
	wrong, _ := alerts.NewCipher(base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{5}, 32)))
	s.Cipher = wrong
	if err = s.VerifyChannels(ctx); !errors.Is(err, ErrUnavailable) {
		t.Fatal("chave incompatível aceita", err)
	}
	yes := true
	if _, err = s.UpdateChannel(ctx, c.ID, nil, &yes, false); !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
	s.Cipher = nil
	if _, err = s.CreateChannel(ctx, channelInput()); !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
}
