// 包 k8s_cross 是一个 CoreDNS 插件，用于演示如何开发 CoreDNS 插件。
//
// 该插件展示了 CoreDNS 插件的基本结构和关键组件。
package k8s_cross

import (
	"context"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/miekg/dns"
)

// 定义一个包含插件名称的日志记录器。这样我们就可以直接使用 log.Info 和
// 其他相关方法进行日志记录。
var log = clog.NewWithPlugin("k8s_cross")

// K8sCross 是 k8s_cross 插件的主要结构体，用于处理 DNS 请求。
type K8sCross struct {
	Next plugin.Handler
}

// ServeDNS 实现了 plugin.Handler 接口。这是插件处理 DNS 请求的入口点。
// 参数:
// - ctx: 请求上下文，包含请求相关信息
// - w: DNS 响应写入器，用于发送响应给客户端
// - r: DNS 请求消息
// 返回值:
// - int: DNS 响应码
// - error: 处理过程中的错误
func (e K8sCross) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	// 记录接收到请求的调试信息
	log.Debug("Received DNS request")

	// 创建自定义响应写入器来处理响应
	pw := NewResponsePrinter(w)

	// 增加请求计数指标
	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

	// 调用插件链中的下一个处理器
	return plugin.NextOrFailure(e.Name(), e.Next, ctx, pw, r)
}

// Name 实现了 Handler 接口，返回插件名称。
func (e K8sCross) Name() string { return "k8s_cross" }

// ResponsePrinter 包装了一个 dns.ResponseWriter，并在调用 WriteMsg 时向标准输出写入 "example"。
type ResponsePrinter struct {
	dns.ResponseWriter
}

// NewResponsePrinter 创建并返回一个新的 ResponsePrinter 实例。
// 参数:
// - w: 原始的 dns.ResponseWriter
// 返回值:
// - *ResponsePrinter: 新创建的 ResponsePrinter 实例
func NewResponsePrinter(w dns.ResponseWriter) *ResponsePrinter {
	return &ResponsePrinter{ResponseWriter: w}
}

// WriteMsg 调用底层 ResponseWriter 的 WriteMsg 方法，并向标准输出打印 "example"。
func (r *ResponsePrinter) WriteMsg(res *dns.Msg) error {
	log.Info("example")
	return r.ResponseWriter.WriteMsg(res)
}
