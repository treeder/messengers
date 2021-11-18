package telegram

import (
	"context"
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/treeder/messengers"
)

type InMsg struct {
	Update *tgbotapi.Update // could drop Msg below and just have this, basically a dupe now
	Msg    *tgbotapi.Message
	cmd    string
	split  []string
}

func NewInMsg(ctx context.Context, update *tgbotapi.Update) *InMsg {
	msg := update.Message
	cmd, tsplit := messengers.ParseCommand(ctx, msg.Text)
	msgI := &InMsg{update, msg, cmd, tsplit}
	return msgI
}

func (m *InMsg) ID() string {
	return strconv.FormatInt(int64(m.Msg.MessageID), 10)
}

func (m *InMsg) Command() string {
	return m.cmd
}

func (m *InMsg) ChatID() string {
	return strconv.FormatInt(int64(m.Msg.Chat.ID), 10)
}

func (m *InMsg) ThreadID() string {
	return ""
}

func (m *InMsg) TeamID() string {
	return ""
}

func (m *InMsg) FromID() string {
	return strconv.FormatInt(int64(m.Msg.From.ID), 10)
}
func (m *InMsg) FromUsername() string {
	return m.Msg.From.UserName
}

func (m *InMsg) FullText() string {
	return m.Msg.Text
}
func (m *InMsg) Cmd() string {
	return m.Command()
}
func (m *InMsg) Split() []string {
	return m.split
}
func (m *InMsg) IsPrivate() bool {
	return m.Msg.Chat.IsPrivate()
}

func (m *InMsg) Mention() string {
	return messengers.MarkdownEscape(fmt.Sprintf("@%v", m.FromUsername()))
}

func (m *InMsg) ReplyToMsgID() string {
	if m.Msg.ReplyToMessage == nil {
		return ""
	}
	mid := m.Msg.ReplyToMessage.MessageID
	if mid == 0 {
		return ""
	}
	return strconv.FormatInt(int64(mid), 10)
}

func (m *InMsg) ReplyToMsg() messengers.Message {
	m2 := m.Msg.ReplyToMessage
	if m2 == nil {
		return nil
	}
	msgI := &InMsg{Msg: m2}
	return msgI
}

// Raw returns *tgbotapi.Message
func (m *InMsg) Raw() interface{} {
	return m.Update
}
