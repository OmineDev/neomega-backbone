package neomega_backbone

type DynamicComponentConfig interface {
	Upgrade(any) error
	Configs() any
}

// Component 描述了组件应该具有的接口
// 顺序 &Component{} -> .Init(ComponentConfig) -> Activate() -> Stop()
// 每个 Activate 工作在一个独立的 goroutine 下
type DynamicComponent interface {
	Init(cfg DynamicComponentConfig, storage StorageAndPathAccess)
	Inject(frame ExtendOmega)
	BeforeActivate() error
	Activate()
	// 尽量不要依赖 Stop, 因为用户往往会选择直接关闭程序而非 ctrl c, 此时 Stop 没有机会被调用
	// Stop() error
}

type ChallengeFn func(challenge string) (response string)
type DynamicComponentFactory func(name string, fn ChallengeFn) DynamicComponent

type BasicDynamicComponent struct {
	Config DynamicComponentConfig
	Frame  ExtendOmega
}

func (bc *BasicDynamicComponent) Init(cfg DynamicComponentConfig, storage StorageAndPathAccess) {
	bc.Config = cfg
}

func (bc *BasicDynamicComponent) Inject(frame ExtendOmega) {
	bc.Frame = frame
}

func (bc *BasicDynamicComponent) BeforeActivate() error {
	return nil
}

func (bc *BasicDynamicComponent) Activate() {
}

// func (bc *BasicDynamicComponent) Stop() error {
// 	return nil
// }
