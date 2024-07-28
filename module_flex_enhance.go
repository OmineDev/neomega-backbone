package neomega_backbone

import (
	"github.com/OmineDev/neomega-core/nodes/defines"
	"github.com/OmineDev/neomega-core/utils/async_wrapper"
)

type FlexEnhance interface {
	defines.FundamentalNode
	defines.KVDataNode
	defines.RolesNode
	defines.TimeLockNode
	// args should always be a json string
	SoftCallOmitResult(cmd, args string)
	// args should always be a json string
	SoftCall(cmd, args string) async_wrapper.AsyncResult[string]
	// e.g. if we what to add a new cmd ban <player> <reason> <time>
	// we can register a new cmd like this:
	// RegisterSoftAPI("ban", onCmd)
	// where onCmd is a func(args string) output string
	// and args is a string like "<player> <reason> <time>"
	// args and ret should always be a json string
	RegisterSoftAPI(cmd string) async_wrapper.AsyncAPISetHandler[string, string]

	// e.g. if we want to add a new topic "player_banned"
	// it can be used like this:
	// SoftPublish("player_banned", "player1")
	// SoftPublish("player_banned","player2")
	// bannedPlayer := SoftListen("player_banned")
	// player1 := <-bannedPlayer
	// player2 := <-bannedPlayer
	SoftPublish(topic string, msg string)
	SoftListen(topic string, nonBlockingMsgHandleFn func(string))

	SoftGet(key string) (val string, found bool)
	SoftSet(key string, val string)

	// cannot get/set in other process
	InProcessGet(key any) (val any, found bool)
	InProcessSet(key any, val any)
	InProcessCompareAndDelete(key any, old any) (deleted bool)
	InProcessCompareAndSwap(key any, old any, new any) bool
	InProcessDelete(key any)
	InProcessLoadAndDelete(key any) (value any, loaded bool)
	InProcessLoadOrStore(key any, value any) (actual any, loaded bool)
	InProcessRange(f func(key any, value any) bool)
	InProcessSwap(key any, value any) (previous any, loaded bool)

	InProcessListen(topic string, onMsg func(any), newGoroutine bool)
	InProcessPublish(topic string, msg any)
	RegInProcessAPI(apiName string) async_wrapper.AsyncAPISetHandler[any, any]
	InProcessCallAPI(apiName string, args any) async_wrapper.AsyncResult[any]
	InProcessCallAPIOmitResponse(apiName string, args any)
}
