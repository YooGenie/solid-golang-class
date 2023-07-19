package storage_providers

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	EDP_ES_STORAGE_PROVIDER_WRITE_TOTAL      = "edp_es_storage_provider_write_total"
	EDP_ES_STORAGE_PROVIDER_WRITE_TOTAL_HELP = "the number of messages that elasticsearch storage provider wrote in total"
)

var (
	esWriteTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: EDP_ES_STORAGE_PROVIDER_WRITE_TOTAL,
		Help: EDP_ES_STORAGE_PROVIDER_WRITE_TOTAL_HELP},
	)
)

func init() {
	prometheus.Register(esWriteTotal)
}
