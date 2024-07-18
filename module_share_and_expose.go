package neomega_backbone

import (
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/utils/async_wrapper"
)

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

// type TerminalInputHandler interface {
// 	AsyncCallBackInGoRoutine(cb func(nextInput string))
// 	BlockGet() string
// }

type Printer interface {
	Write(string)
	Print(...interface{})
	Printf(format string, a ...interface{})
	Println(...interface{})
	Printfln(format string, a ...interface{})
}

type MultiOutDst struct {
	// print to terminal only
	Printer
	// print to terminal only
	Terminal Printer
	// print to log only
	Log Printer
	// print to terminal and log
	TerminalAndLog Printer
	// Debug (lowest level), whether it will be on terminal or log decide on frame config
	// default will not shown in terminal or log
	Debug Printer
	// Info (normal level), whether it will be on terminal or log decide on frame config
	// default will shown in terminal but not log
	Info Printer
	// SuccessInfo (normal level), whether it will be on terminal or log decide on frame config
	// default will shown in terminal but not log
	Success Printer
	// Warning (high priority), whether it will be on terminal or log decide on frame config
	// default will shown in both terminal and log
	Warning Printer
	// Error (high priority), whether it will be on terminal or log decide on frame config
	// default will shown in both terminal and log
	Error Printer
}

// 一个特定的模块(e.g. BackendMenu)实现这个接口，读取所有注册的接口，然后在终端显示
type BackendIO interface {
	// 设置终端菜单项， 用于注册
	AddBackendMenuEntry(*BackendMenuEntry)
	// SetOnTerminalInputCallBack(func(string))
	GetTerminalInput() async_wrapper.AsyncResult[string]
	Out() *MultiOutDst
	ForkOut(prefix string, logFile string) *MultiOutDst
}

type BackendIOModule interface {
	BackendIO
	CanPreInit
}
