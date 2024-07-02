package neomega_backbone

import "context"

type CanPreInit interface {
	PreInit(PreInitOmega) error
}

type CanInteractWithOmega interface {
	Init(ctx context.Context, omega ExtendOmega)
}

type HasLifeCycle interface {
	Init(ctx context.Context)
}

type BasicConfig struct {
	Name     string `json:"名称"`
	Source   string `json:"来源"`
	Disabled bool   `json:"是否禁用"`
}

type ConfigWrite interface {
	AddDefaultConfigFile(basicConfig *BasicConfig, fullConfig any, onWriteCallBack func(*BasicConfig, any))
}

type ConfigRead interface {
	GetEnabledConfigBySource(source string) []string
}

type ConfigProvider interface {
	// GetFrameConfigOverride returns the frame config override, can be nil
	// if not nil, it will be used to override/adjust the config
	GetFrameConfigOverride(config any) any
	ConfigRead
	ConfigWrite
	CanPreInit
}
