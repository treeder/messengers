package discord

import (
	"context"

	"github.com/bsdlp/discord-interactions-go/interactions"
	"github.com/treeder/messengers"
)

type InMsg struct {
	Msg   interactions.Data
	cmd   string
	split []string
	ctx   context.Context
}

func (m *InMsg) ID() string {
	return m.Msg.ID
}

func (m *InMsg) Command() string {
	return m.cmd
}
func (m *InMsg) IsSlashCommand() bool {
	return true // always a slash command
}

func (m *InMsg) TeamID() string {
	return m.Msg.GuildID
}
func (m *InMsg) ChatID() string {
	return m.Msg.ChannelID
}

func (m *InMsg) FromID() string {
	return m.Msg.Member.User.ID
}
func (m *InMsg) FromUsername() string {
	return m.Msg.Member.User.Username
}
func (m *InMsg) FullText() string {
	return "NO IDEA" //m.Msg.Content
}
func (m *InMsg) Cmd() string {
	return m.Command()
}
func (m *InMsg) Split() []string {
	return m.split
}
func (m *InMsg) IsPrivate() bool {
	return false
	// var err error
	// if m.channel == nil {
	// 	m.channel, err = m.sess.Channel(m.Msg.ChannelID)
	// 	if err != nil {
	// 		if m.ctx == nil {
	// 			m.ctx = context.TODO()
	// 		}
	// 		gotils.LogBeta(context.TODO(), "error", "discord error getting channel", err)
	// 		return false
	// 	}
	// }
	// m.Msg.ChannelID
	// return m.channel.Type == discordgo.ChannelTypeDM
}

func (m *InMsg) Mention() string {
	return m.Msg.Member.User.Username
}

// ReplyToMsgID always returns "" since Discord doesn't have replies
func (m *InMsg) ReplyToMsgID() string {
	return ""
}

// ReplyToMsg always returns nil since Discord doesn't have replies
func (m *InMsg) ReplyToMsg() messengers.Message {
	return nil
}

func (m *InMsg) ThreadID() string {
	return ""
}

// Raw returns *discordgo.Message
func (m *InMsg) Raw() interface{} {
	return m.Msg
}
