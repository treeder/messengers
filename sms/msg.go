package sms

import (
	"strings"

	"github.com/sfreiberg/gotwilio"
	"github.com/treeder/messengers"
)

type Message struct {
	raw   *gotwilio.SMSWebhook
	cmd   string
	split []string
}

// Cmd the first value in the incoming message, minus the slash
func (m *Message) Cmd() string {
	return m.cmd
}
func (m *Message) Command() string {
	return m.Cmd()
}
func (m *Message) IsSlashCommand() bool {
	return strings.HasPrefix(m.FullText(), "/")
}

func (m *Message) FullText() string {
	return m.raw.Body
}

// Split the incoming command split
func (m *Message) Split() []string {
	return m.split
}

// If a message comes in as a reply, these will be populated. Not all messengers support this.
func (m *Message) ReplyToMsgID() string {
	return ""
}
func (m *Message) ReplyToMsg() messengers.Message {
	return nil
}

func (m *Message) ID() string {
	return m.raw.MessageSid
}

func (m *Message) TeamID() string {
	return ""
}

//ChatID is the room/group/channel/space
func (m *Message) ChatID() string {
	return m.raw.To
}
func (m *Message) FromID() string {
	return m.raw.From
}
func (m *Message) ThreadID() string {
	return ""
}
func (m *Message) FromUsername() string {
	return m.raw.From
}

// IsPrivate says whether this is a private DM with the bot
func (m *Message) IsPrivate() bool {
	// todo: what about group messages?
	return true
}

// Mention spits out the appropriate string to tag the author
func (m *Message) Mention() string {
	return ""
}

// Raw returns the original message from the API client
func (m *Message) Raw() interface{} {
	return m.raw
}
