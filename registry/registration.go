package registry

type Registration struct {
	ServiceName ServiceName
	ServiceURL  string
	// 服务发现
	RequiredServices []ServiceName //当前服务依赖的其他服务，比如grade依赖log，那么log就应该在里面
	ServiceUpdateUrl string        // 更新服务地址，暴露服务地址
	HeartbeatURL     string
}

type ServiceName string

const (
	LogService     = ServiceName("LogService")
	GradingService = ServiceName("GradingService")
	PortalService  = ServiceName("Portald")
)

type patchEntry struct {
	Name ServiceName
	URL  string
}

type patch struct {
	Added   []patchEntry
	Removed []patchEntry
}
