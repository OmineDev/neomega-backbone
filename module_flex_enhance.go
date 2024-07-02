package neomega_backbone

import (
	"time"

	"github.com/OmineDev/neomega-core/nodes/defines"
)

type FlexEnhance interface {
	defines.FundamentalNode
	defines.KVDataNode
	defines.RolesNode
	defines.TimeLockNode
	FlexCmd(cmd, args string)
	// allow blocking in onResp, if timeout<=0 then no timeout
	FlexCmdWithCb(cmd, args string, onResp func(resp string), timeout time.Duration)
	// could be blocking, if timeout<=0 then no timeout
	FlexCmdWithBlockResult(cmd, args string, timeout time.Duration) string
	// e.g. if we what to add a new cmd ban <player> <reason> <time>
	// we can register a new cmd like this:
	// RegisterFlexCmd("ban", onCmd)
	// where onCmd is a func(args string) output string
	// and args is a string like "<player> <reason> <time>"
	RegisterFlexCmd(cmd string, onCmd func(args string) string)

	// e.g. if we want to add a new topic "player_banned"
	// it can be used like this:
	// FlexPublish("player_banned", "player1")
	// FlexPublish("player_banned","player2")
	// bannedPlayer := FlexListen("player_banned")
	// player1 := <-bannedPlayer
	// player2 := <-bannedPlayer
	FlexPublish(topic string, msg string)
	FlexListen(topic string, onMsg func(string))
}
