package neomega_backbone

import "github.com/OmineDev/qq-bot-helper/packet"

type DefaultCQMessageCb func(source, name, message string)

type CQHTTPAccess interface {
	RegisterPacketNoBlockCB(cb func(pk packet.CQPacket, data []byte))
	SendGroupMessage(groupID int64, message string, onCb func(ok bool, msgID int64))
	GetGroupMember(groupID int64, onCb func(packet.GroupMemberCards))
	GetGuildList(onCb func(guilds packet.GuildList))
	GetGuildChannels(guildID string, onCb func(channels packet.GuildChannels))
	GetGuildMemberProfile(guildID, userID string, onCB func(packet.GuildMemberProfile))
	SendGuildMessage(guildID, channelID string, message string, onCb func(ok bool, msgID int64))
	SendPrivateMessage(userID int64, message string, onCb func(ok bool, msgID int64))
	// send message to default target,
	// what is the default target and how should the message convert, also how to send to default target is decided by cqhttp.lua
	SendToDefault(message string)

	// listen message from default target (name, message, source string) source=好友:昵称orQQ号/群聊:群号/频道:频道名:聊天室
	// what is the default target(could be private msg, group msg, guild msg, and even a combination of them) is decided by cqhttp.lua
	// also, how to get name of each message is also decided by cqhttp.lua
	OnDefaultMessage(cb DefaultCQMessageCb)

	// send message to target, target has same format as source in OnDefaultMessage, decided by cqhttp.lua
	// so, you can reply to a specific target by letting target=source
	// when target="", it means send to default target (SendToDefault)
	SendTo(target, message string)
}

type CQHTTP interface {
	CQHTTPAccess
	CanPreInit
}
