package neomega_backbone

import (
	"github.com/OmineDev/neomega-core/nodes/defines"
	"github.com/OmineDev/neomega-core/utils/async_wrapper"
)

type CanSetArg interface {
	WithArg(key string, arg any) CanSetArg
	Launch()
}

type CanSetData interface {
	WithJsonStrData(jsonStrData string)
	WithJsonBytesData(jsonBytesData []byte)
	WithJsonableAny(jsonableData any)
	WithArg(key string, arg any) CanSetArg
}

type CanGetData interface {
	RawJsonStr() string
	RawJsonBytes() []byte
	AsMap() (map[string]any, error)
	Bind(target any) error
	GetValue(key string) any
	TakeValue(key string, value any) CanGetData
}

type CanSetArgThenResult interface {
	WithArg(key string, arg any) CanSetArgThenResult
	Launch() async_wrapper.AsyncResult[CanGetData]
}

type CanSetDataThenResult interface {
	WithJsonStrData(jsonStrData string) async_wrapper.AsyncResult[CanGetData]
	WithJsonBytesData(jsonBytesData []byte) async_wrapper.AsyncResult[CanGetData]
	WithJsonableAny(jsonableData any) async_wrapper.AsyncResult[CanGetData]
	WithArg(key string, arg any) CanSetArgThenResult
}

type FlexEnhance interface {
	defines.FundamentalNode
	defines.KVDataNode
	defines.RolesNode
	defines.TimeLockNode
	// args should always be a json string
	SoftCallOmitResult(cmd string) CanSetData
	// args should always be a json string
	SoftCall(cmd string) CanSetDataThenResult
	// e.g. if we what to add a new cmd ban <player> <reason> <time>
	// we can register a new cmd like this:
	// RegSoftAPI("ban", onCmd)
	// where onCmd is a func(args string) output string
	// and args is a string like "<player> <reason> <time>"
	// args and ret should always be a json bytes
	RegSoftAPI(cmd string) async_wrapper.AsyncAPISetHandler[CanGetData, []byte]

	// e.g. if we want to add a new topic "player_banned"
	// it can be used like this:
	// SoftPublish("player_banned", "player1")
	// SoftPublish("player_banned","player2")
	// bannedPlayer := SoftListen("player_banned")
	// player1 := <-bannedPlayer
	// player2 := <-bannedPlayer
	SoftPublish(topic string) CanSetData
	SoftListen(topic string, nonBlockingMsgHandleFn func(CanGetData))

	SoftGet(key string) (val CanGetData, found bool)
	SoftSet(key string) CanSetData

	// cannot get/set in other process
	InProcessGet(key string) (val any, found bool)
	InProcessSet(key string, val any)
	InProcessCompareAndDelete(key string, old any) (deleted bool)
	InProcessCompareAndSwap(key string, old any, new any) bool
	InProcessDelete(key string)
	InProcessLoadAndDelete(key string) (value any, loaded bool)
	InProcessLoadOrStore(key string, value any) (actual any, loaded bool)
	InProcessRange(f func(key string, value any) bool)
	InProcessSwap(key string, value any) (previous any, loaded bool)

	InProcessListen(topic string, onMsg func(any), newGoroutine bool)
	InProcessPublish(topic string, msg any)
	RegInProcessAPI(apiName string) async_wrapper.AsyncAPISetHandler[any, any]
	InProcessCallAPI(apiName string, args any) async_wrapper.AsyncResult[any]
	InProcessCallAPIOmitResponse(apiName string, args any)
}
