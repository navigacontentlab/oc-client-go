package oc

import (
	"context"
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	StatusCodes  *prometheus.CounterVec
	Duration     *prometheus.HistogramVec
	OrgExtractor func(ctx context.Context) string
}

func DefaultOrganisationExtractor(_ context.Context) string {
	return "unknown"
}

func NewMetrics(reg prometheus.Registerer, orgExtractor func(ctx context.Context) string) (*Metrics, error) {
	//nolint:promlinter
	statusCodes := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "oc_status_codes",
			Help: "HTTP status codes returned from OC.",
		},
		[]string{"path", "status", "organisation"},
	)
	if err := reg.Register(statusCodes); err != nil {
		return nil, fmt.Errorf("failed to register metric: %w", err)
	}

	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "oc_duration",
		Help:    "Duration for an oc call.",
		Buckets: prometheus.ExponentialBuckets(5, 2, 15),
	}, []string{"path", "organisation"})
	if err := reg.Register(duration); err != nil {
		return nil, fmt.Errorf("failed to register metric: %w", err)
	}

	if orgExtractor == nil {
		orgExtractor = DefaultOrganisationExtractor
	}

	return &Metrics{
		StatusCodes:  statusCodes,
		Duration:     duration,
		OrgExtractor: orgExtractor,
	}, nil
}

func (m *Metrics) incStatusCode(ctx context.Context, path string, status int) {
	organisation := m.OrgExtractor(ctx)

	m.StatusCodes.WithLabelValues(path, strconv.Itoa(status), organisation).Inc()
}

func (m *Metrics) addDuration(ctx context.Context, path string, milliseconds float64) {
	organisation := m.OrgExtractor(ctx)

	m.Duration.WithLabelValues(path, organisation).Observe(milliseconds)
}
