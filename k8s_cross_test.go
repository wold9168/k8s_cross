package k8s_cross

import (
	"bytes"
	"context"
	golog "log"
	"strings"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"

	"github.com/miekg/dns"
)

// TestK8sCross 测试 k8s_cross 插件的核心功能
func TestK8sCross(t *testing.T) {
	// 创建一个新的 K8sCross 插件实例，使用 test.ErrorHandler 作为下一个插件处理器
	x := K8sCross{Next: test.ErrorHandler()}

	// 创建一个缓冲区来捕获日志输出，而不是直接输出到标准输出
	b := &bytes.Buffer{}
	golog.SetOutput(b)

	ctx := context.TODO()
	r := new(dns.Msg)
	r.SetQuestion("example.org.", dns.TypeA)
	// 创建一个 Recorder 来模拟 DNS 响应写入器
	rec := dnstest.NewRecorder(&test.ResponseWriter{})

	// 调用插件的 ServeDNS 方法处理请求
	x.ServeDNS(ctx, rec, r)
	
	// 验证插件是否正确输出了预期的日志信息
	if a := b.String(); !strings.Contains(a, "[INFO] plugin/k8s_cross: Processing request in k8s_cross plugin") {
		t.Errorf("Expected log message not found. Got: %s", a)
	}
}
