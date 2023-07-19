package kafka

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	EDP_KAFKA_CONSUMER_READ_TOTAL      = "edp_kafka_consumer_read_total"
	EDP_KAFKA_CONSUMER_READ_TOTAL_HELP = "the number of messages that kafka consumer reads in total"
)

var (
	ConsumerReadTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: EDP_KAFKA_CONSUMER_READ_TOTAL,
		Help: EDP_KAFKA_CONSUMER_READ_TOTAL_HELP},
	)
)

func init() {
	prometheus.Register(ConsumerReadTotal)
}
