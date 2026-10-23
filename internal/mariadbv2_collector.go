package internal

import (
	"github.com/prometheus/client_golang/prometheus"
)

type MariaDBV2Collector struct {
	clusterMetric  *prometheus.GaugeVec
	descFetchError *prometheus.Desc
}

func NewMariaDBV2Collector() *MariaDBV2Collector {
	return &MariaDBV2Collector{
		clusterMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ionos_mariadb_v2_cluster",
			Help: "MariaDBV2 cluster exposed by its id, name and version",
		}, []string{"id", "name", "version"}),
		descFetchError: prometheus.NewDesc(
			"ionos_mariadb_v2_fetch_error",
			"Error during MariaDB V2 metrics fetch",
			nil,
			nil,
		),
	}
}

func (collector *MariaDBV2Collector) Describe(ch chan<- *prometheus.Desc) {
	collector.clusterMetric.Describe(ch)
	ch <- collector.descFetchError
}

func (collector *MariaDBV2Collector) Collect(ch chan<- prometheus.Metric) {
	mariadbV2Mutex.RLock()
	defer mariadbV2Mutex.RUnlock()

	collector.clusterMetric.Reset()
	for _, cluster := range mariadbV2Clusters {
		collector.clusterMetric.WithLabelValues(cluster.ID, cluster.Name, cluster.Version).Set(1)
	}
	collector.clusterMetric.Collect(ch)

	ch <- prometheus.MustNewConstMetric(
		collector.descFetchError,
		prometheus.GaugeValue, mariadbV2FetchError,
	)
}
