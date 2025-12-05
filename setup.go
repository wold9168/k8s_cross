package k8s_cross

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

// init 注册这个插件。
func init() { plugin.Register("k8s_cross", setup) }

// setup 是当配置解析器看到 "k8s_cross" 标记时调用的函数。setup 负责
// 解析 k8s_cross 插件可能有的任何额外选项。这个函数看到的第一个标记是 "k8s_cross"。
func setup(c *caddy.Controller) error {
	c.Next() // 忽略 "k8s_cross" 并给我们下一个标记。
	if c.NextArg() {
		// 如果有另一个标记，返回错误，因为我们没有任何配置。
		// 从这个 setup 函数返回的任何错误都应该用 plugin.Error 包装，这样我们
		// 可以为用户呈现一个稍微好一点的错误消息。
		return plugin.Error("k8s_cross", c.ArgErr())
	}

	// 将插件添加到 CoreDNS，以便服务器可以在其插件链中使用它。
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return K8sCross{Next: next}
	})

	// 一切正常，返回 nil 错误。
	return nil
}
