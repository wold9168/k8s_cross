package k8s_cross

// Ready 实现了 ready.Readiness 接口，用于报告插件是否已准备好处理查询。
// 当此方法返回 true 时，CoreDNS 认为插件已就绪，此后不再检查。
// 对于 k8s_cross 插件，由于不需要特殊初始化，始终返回 true。
func (e K8sCross) Ready() bool { return true }
