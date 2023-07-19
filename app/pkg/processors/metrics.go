package processors

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	EDP_KAFKA_PROCESSOR_PROCESS_TOTAL      = "edp_kafka_processor_process_total"
	EDP_KAFKA_PROCESSOR_PROCESS_TOTAL_HELP = "the number of messages that kafka default processor processed in total"
)

var (
	KafkaProcessTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: EDP_KAFKA_PROCESSOR_PROCESS_TOTAL,
		Help: EDP_KAFKA_PROCESSOR_PROCESS_TOTAL_HELP},
	)
)

func init() {
	prometheus.Register(KafkaProcessTotal)
}
