package middlewares

import (
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

type WebMonitoring interface {
	MonitorWebRequest() gin.HandlerFunc
	Metrics() gin.HandlerFunc
}

type PrometheusMonitoring struct {
	httpRequestTotal    *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	httpResponseStatus  *prometheus.CounterVec
	goRoutines          prometheus.Gauge
	dbConnections       prometheus.Gauge
	cpuUsage            prometheus.Gauge
	memoryUsage         prometheus.Gauge
	networkRxBytes      prometheus.Counter
	networkTxBytes      prometheus.Counter
}

func NewPrometheusMonitorWebRequest() WebMonitoring {
	m := &PrometheusMonitoring{
		httpRequestTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_request_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "http_request_duration_seconds",
				Help: "Duration of HTTP requests",
			},
			[]string{"method", "path"},
		),
		httpResponseStatus: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_response_status",
				Help: "Status of HTTP response",
			},
			[]string{"method", "path", "status"},
		),
		goRoutines: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "app_goroutines",
				Help: "Number of goroutines in the application",
			},
		),
		dbConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections",
				Help: "Number of database connections",
			},
		),
		cpuUsage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "cpu_usage_percent",
				Help: "CPU usage in percent",
			},
		),
		memoryUsage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "memory_usage_bytes",
				Help: "Memory usage in bytes",
			},
		),
		networkRxBytes: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "network_receive_bytes_total",
				Help: "Total number of bytes received",
			},
		),
		networkTxBytes: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "network_transmit_bytes_total",
				Help: "Total number of bytes transmitted",
			},
		),
	}
	prometheus.MustRegister(m.httpRequestTotal, m.httpRequestDuration, m.httpResponseStatus, m.goRoutines, m.dbConnections, m.cpuUsage, m.memoryUsage, m.networkRxBytes, m.networkTxBytes)

	// Delete metrics every minute
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			<-ticker.C
			m.httpRequestTotal.Reset()
			m.httpRequestDuration.Reset()
			m.httpResponseStatus.Reset()
		}
	}()

	// Regular metrics update
	go m.updateMetrics()

	return m
}

func (m *PrometheusMonitoring) MonitorWebRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		m.httpRequestDuration.WithLabelValues(c.Request.Method, c.Request.URL.Path).Observe(duration.Seconds())
		m.httpRequestTotal.WithLabelValues(c.Request.Method, c.Request.URL.Path, strconv.Itoa(c.Writer.Status())).Inc()
		// get status code
		status := c.Writer.Status()
		m.httpResponseStatus.WithLabelValues(c.Request.Method, c.Request.URL.Path, strconv.Itoa(status)).Inc()
	}
}

func (m *PrometheusMonitoring) Metrics() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

func (p *PrometheusMonitoring) updateMetrics() {
	ticker := time.NewTicker(15 * time.Second)
	for range ticker.C {
		// Update Go routine count
		p.goRoutines.Set(float64(runtime.NumGoroutine()))

		// Update CPU usage
		if cpuPercent, err := cpu.Percent(0, false); err == nil {
			p.cpuUsage.Set(cpuPercent[0])
		}

		// Update memory usage
		if memStats, err := mem.VirtualMemory(); err == nil {
			p.memoryUsage.Set(float64(memStats.Used))
		}

		// Update network statistics
		if netStats, err := net.IOCounters(false); err == nil {
			p.networkRxBytes.Add(float64(netStats[0].BytesRecv))
			p.networkTxBytes.Add(float64(netStats[0].BytesSent))
		}
	}
}
