package k8s_cross

import (
	"testing"

	"github.com/coredns/caddy"
)

// TestSetup 测试 k8s_cross 插件的配置解析功能。
// 验证插件能否正确处理有效的配置以及拒绝无效的配置。
func TestSetup(t *testing.T) {
	// 测试有效的配置（无参数）
	c := caddy.NewTestController("dns", `k8s_cross`)
	if err := setup(c); err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	// 测试无效的配置（带参数）
	c = caddy.NewTestController("dns", `k8s_cross more`)
	if err := setup(c); err == nil {
		t.Fatalf("Expected error, but got none")
	}
}
