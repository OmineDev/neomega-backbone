package neomega_backbone

import (
	"utils/sync_wrapper"

	"github.com/OmineDev/neomega-core/neomega"
)

// FrameLevelShare aims to provide a common interface for data/api share between different components/module/loader
type FrameLevelShare interface {
	// both data and api share are supported
	Share(key string, value any)
	Get(key string) (value any, ok bool)
	GetOmegaUserInfo() map[string]string
}

type MenuEntry struct {
	// 触发词, 不可为空
	Triggers []string // e.g ["tp", "前往", "去坐标"]
	// 参数提示
	ArgumentHint string // e.g "[x] [y] [z]"
	Usage        string // e.g. "前往指定坐标 [x] [y] [z]"
}

// 游戏内菜单项，被聊天触发
type GameMenuEntry struct {
	// 触发条件和基本信息
	MenuEntry
	// 触发后的回调函数
	OnTrigCallBack func(chat *neomega.GameChat)
}

// 一个特定的模块(e.g. GameMenu)实现这个接口，读取所有注册的接口，然后在游戏内显示
type GameMenuSetter interface {
	// 设置游戏内菜单项， 用于注册
	AddGameMenuEntry(*GameMenuEntry)
}

type GameMenuModule interface {
	GameMenuSetter
	CanPreInit
}

// 终端菜单项，被命令触发
type BackendMenuEntry struct {
	MenuEntry
	OnTrigCallBack func(cmds []string)
}

type TerminalInputHandler interface {
	AsyncCallBackInGoRoutine(cb func(nextInput string))
	BlockGet() string
}

// 一个特定的模块(e.g. BackendMenu)实现这个接口，读取所有注册的接口，然后在终端显示
type BackendIO interface {
	// 设置终端菜单项， 用于注册
	AddBackendMenuEntry(*BackendMenuEntry)
	SetOnTerminalInputCallBack(func(string))
	GetTerminalInput() TerminalInputHandler
	ColorPrint(s string)
}

type BackendIOModule interface {
	BackendIO
	CanPreInit
}

type StringKVSyncMapShare interface {
	GetSharedMap() *sync_wrapper.SyncKVEnhancedMap[string, string]
}
