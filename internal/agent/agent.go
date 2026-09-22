package agent

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/fcconsultoria/kmind-agent-docker/internal/buffer"
	"github.com/fcconsultoria/kmind-agent-docker/internal/config"
	dockerclient "github.com/fcconsultoria/kmind-agent-docker/internal/docker"
	"github.com/fcconsultoria/kmind-agent-docker/internal/identity"
	"github.com/fcconsultoria/kmind-agent-docker/internal/status"
	"github.com/fcconsultoria/kmind-agent-docker/internal/transport"
	collectormetricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/proto"
)

type Agent struct {
	config    config.Config
	identity  identity.Identity
	version   string
	logger    *slog.Logger
	docker    dockerclient.Reader
	status    *status.Client
	transport *transport.Client
	buffer    *buffer.Queue
	active    atomic.Bool
}

func New(cfg config.Config, installation identity.Identity, version string, logger *slog.Logger) (*Agent, error) {
	docker, err := dockerclient.New(cfg.Docker.Socket)
	if err != nil {
		return nil, err
	}
	return &Agent{config: cfg, identity: installation, version: version, logger: logger, docker: docker, status: status.New(cfg.Kmind.StatusEndpoint, cfg.Kmind.APIKey, cfg.Limits.RequestTimeout.Value()), transport: transport.New(cfg.Kmind.Endpoint, cfg.Kmind.APIKey, cfg.Limits.RequestTimeout.Value()), buffer: buffer.New(cfg.Limits.MemoryBufferBytes)}, nil
}
func (a *Agent) Run(ctx context.Context) error {
	go a.statusLoop(ctx)
	ticker := time.NewTicker(a.config.Collection.MetricsInterval.Value())
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			a.active.Store(false)
			a.buffer.Clear()
			return ctx.Err()
		case <-ticker.C:
			if a.active.Load() {
				a.heartbeat(ctx)
			}
		}
	}
}
func (a *Agent) statusLoop(ctx context.Context) {
	delay := time.Duration(0)
	for {
		if delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
		state, err := a.status.Check(ctx)
		if err != nil {
			a.active.Store(false)
			a.buffer.Clear()
			a.logger.Warn("agent status check failed; telemetry paused", "error", err)
			delay = time.Minute
			continue
		}
		a.active.Store(state.Active)
		if !state.Active {
			a.buffer.Clear()
			delay = max(state.CheckAfter, 5*time.Minute)
			a.logger.Info("agent key is inactive; telemetry paused")
		} else {
			delay = state.CheckAfter
		}
	}
}
func (a *Agent) heartbeat(ctx context.Context) {
	operationCtx, cancel := context.WithTimeout(ctx, a.config.Limits.RequestTimeout.Value())
	defer cancel()
	if err := a.docker.Ping(operationCtx); err != nil {
		a.logger.Warn("Docker API is unavailable", "error", err)
		return
	}
	payload, err := a.heartbeatPayload()
	if err != nil {
		return
	}
	if err := a.transport.Send(operationCtx, "metrics", payload); err != nil {
		a.logger.Warn("send heartbeat", "error", err)
	}
}
func (a *Agent) heartbeatPayload() ([]byte, error) {
	now := uint64(time.Now().UnixNano())
	metric := &metricsv1.Metric{
		Name: "kmind_agent_active",
		Data: &metricsv1.Metric_Gauge{Gauge: &metricsv1.Gauge{DataPoints: []*metricsv1.NumberDataPoint{{
			TimeUnixNano: now,
			Value:        &metricsv1.NumberDataPoint_AsDouble{AsDouble: 1},
		}}}},
	}
	request := &collectormetricsv1.ExportMetricsServiceRequest{ResourceMetrics: []*metricsv1.ResourceMetrics{{
		Resource: &resourcev1.Resource{Attributes: a.identity.ResourceAttributes(
			a.config.Kmind.ServiceName, a.config.Kmind.ApplicationName, a.version,
		)},
		ScopeMetrics: []*metricsv1.ScopeMetrics{{
			Scope:   &commonv1.InstrumentationScope{Name: "kmind-agent-docker"},
			Metrics: []*metricsv1.Metric{metric},
		}},
	}}}
	return proto.Marshal(request)
}
