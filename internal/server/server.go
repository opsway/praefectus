package server

import (
	"net"
	"net/http"
	"strconv"

	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/opsway/praefectus/internal/config"
	"github.com/opsway/praefectus/internal/metrics"
)

type Server struct {
	config  *config.Config
	metrics *metrics.Metrics
}

func New(cfg *config.Config, m *metrics.Metrics) *Server {
	return &Server{
		config:  cfg,
		metrics: m,
	}
}

func (s *Server) Start() {
	addr := net.JoinHostPort(s.config.Server.Host, strconv.Itoa(s.config.Server.Port))
	health := healthcheck.NewHandler()

	http.HandleFunc("/ready", health.ReadyEndpoint)
	http.HandleFunc("/live", health.LiveEndpoint)
	http.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(addr, nil))
}
