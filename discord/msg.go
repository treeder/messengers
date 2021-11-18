package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/treeder/gotils/v2"
	"github.com/treeder/messengers"
)

type InMsg struct {
	Msg     *discordgo.Message
	sess    *discordgo.Session
	cmd     string
	split   []string
	channel *discordgo.Channel
	ctx     context.Context
	embed   *discordgo.MessageEmbed
}

func (m *InMsg) ID() string {
	return m.Msg.ID
}

func (m *InMsg) Command() string {
	return m.cmd
}
func (m *InMsg) TeamID() string {
	return m.Msg.GuildID
}
func (m *InMsg) ChatID() string {
	return m.Msg.ChannelID
}

func (m *InMsg) FromID() string {
	return m.Msg.Author.ID
}
func (m *InMsg) FromUsername() string {
	return m.Msg.Author.String()
}
func (m *InMsg) FullText() string {
	return m.Msg.Content
}
func (m *InMsg) Cmd() string {
	return m.Command()
}
func (m *InMsg) Split() []string {
	return m.split
}
func (m *InMsg) IsPrivate() bool {
	var err error
	if m.channel == nil {
		m.channel, err = m.sess.Channel(m.Msg.ChannelID)
		if err != nil {
			if m.ctx == nil {
				m.ctx = context.TODO()
			}
			gotils.LogBeta(context.TODO(), "error", "discord error getting channel", err)
			return false
		}
	}
	return m.channel.Type == discordgo.ChannelTypeDM
}

func (m *InMsg) Mention() string {
	return m.Msg.Author.Mention()
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
