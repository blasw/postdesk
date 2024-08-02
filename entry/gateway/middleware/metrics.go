package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpReqTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_req_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"endpoint"},
	)
)

func init() {
	prometheus.MustRegister(httpReqTotal)
}

func AddMetric(endpointName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		httpReqTotal.WithLabelValues(endpointName).Inc()
		c.Next()
	}
}
