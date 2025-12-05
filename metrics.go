package k8s_cross

import (
	"sync"

	"github.com/coredns/coredns/plugin"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// requestCount 导出一个 prometheus 指标，用于统计 k8s_cross 插件处理的请求数量。
// 指标名称遵循 CoreDNS 插件的标准命名规范：coredns_{plugin_name}_request_count_total
var requestCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: plugin.Namespace, // 默认为 "coredns"
	Subsystem: "k8s_cross",      // 插件名称，用作子系统名称
	Name:      "request_count_total",
	Help:      "Counter of DNS requests translated by k8s_cross.",
}, []string{"server"}) // 按服务器标签区分不同实例的指标

var once sync.Once
