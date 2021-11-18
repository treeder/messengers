package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/treeder/gotils/v2"
	"github.com/treeder/messengers"
)

type Inner interface {
	User() string
	Channel() string
	ChannelType() string
	ThreadTs() string
	Text() string
}

type MentionEventWrapper struct {
	ev *slackevents.AppMentionEvent
}

func (w *MentionEventWrapper) User() string {
	return w.ev.User
}

func (w *MentionEventWrapper) Channel() string {
	return w.ev.Channel
}

func (w *MentionEventWrapper) ChannelType() string {
	return "channel" // not sure if this is correct?
}

func (w *MentionEventWrapper) ThreadTs() string {
	return w.ev.ThreadTimeStamp
}

func (w *MentionEventWrapper) Text() string {
	return w.ev.Text
}

type MessageEventWrapper struct {
	ev *slackevents.MessageEvent
}

func (w *MessageEventWrapper) User() string {
	return w.ev.User
}

func (w *MessageEventWrapper) Channel() string {
	return w.ev.Channel
}

func (w *MessageEventWrapper) ChannelType() string {
	return w.ev.ChannelType
}

func (w *MessageEventWrapper) ThreadTs() string {
	return w.ev.ThreadTimeStamp
}

func (w *MessageEventWrapper) Text() string {
	return w.ev.Text
}

type Msg struct {
	mess  *SlackMessenger
	ctx   context.Context
	event *slackevents.EventsAPICallbackEvent
	inner Inner

	cmd   string
	split []string

	userInfo *slack.User

	chatID string
	// postedID is the message timestamp (aka, the msg ID in slack), to be used for replies
	postedID        string
	postedChannelID string
}

func NewMsg(ctx context.Context, mess *SlackMessenger, event *slackevents.EventsAPICallbackEvent, inner slackevents.EventsAPIInnerEvent) *Msg {
	msg := &Msg{
		mess:  mess,
		ctx:   ctx,
		event: event,
	}
	switch ev := inner.Data.(type) {
	case *slackevents.AppMentionEvent:
		msg.inner = &MentionEventWrapper{ev}
	case *slackevents.MessageEvent:
		msg.inner = &MessageEventWrapper{ev}
	}

	text := msg.inner.Text()
	// strip mention, comes in as: "<@ULUNHUPE0> balance"
	// fmt.Println("text before:", text)
	i1 := strings.Index(text, "<@")
	// fmt.Println("i1:", i1)
	if i1 != -1 {
		t1 := text[0:i1]
		// fmt.Println("t1:", t1)
		i2 := strings.Index(text, ">")
		// fmt.Println("i2:", i2)
		text = t1 + text[i2+1:]
	}
	// fmt.Println("text after:", text)
	text = strings.TrimSpace(text)
	msg.cmd, msg.split = messengers.ParseCommand(ctx, text)
	// gotils.LogBeta(ctx, "debug", "MSG: %+v\n", msg)
	return msg
}

func (m *Msg) ID() string {
	if m.postedID != "" {
		return m.postedID
	}
	return m.event.EventID
}

func (m *Msg) Command() string {
	return m.cmd
}

func (m *Msg) ChatID() string {
	if m.chatID != "" {
		return m.chatID
	}
	if m.postedChannelID != "" {
		return m.TeamID() + m.postedChannelID
	}
	return m.inner.Channel()
}

func (m *Msg) ThreadID() string {
	if m.postedID != "" {
		return m.postedID
	}
	return m.inner.ThreadTs()
}

func (m *Msg) TeamID() string {
	return m.event.TeamID
}
func (m *Msg) FromID() string {
	return m.TeamID() + "-" + m.inner.User()
}
func (m *Msg) FromUsername() string {
	if m.userInfo == nil {
		client := m.mess.SlackClient(m.ctx)
		ui, err := client.GetUserInfoContext(m.ctx, m.inner.User())
		if err != nil {
			gotils.LogBeta(m.ctx, "error", "error getting UserProfile for %v from slack: %+v", m.inner.User(), err)
			return "error"
		}
		m.userInfo = ui
	}
	return m.userInfo.Profile.DisplayName
}
func (m *Msg) FullText() string {
	return m.inner.Text()
}
func (m *Msg) Cmd() string {
	return m.Command()
}
func (m *Msg) Split() []string {
	return m.split
}
func (m *Msg) IsPrivate() bool {
	return m.inner.ChannelType() == "im"
}

func (m *Msg) Mention() string {
	return fmt.Sprintf("<@%s>", m.inner.User())
}

// ReplyToMsgID always returns "" since Discord doesn't have replies
func (m *Msg) ReplyToMsgID() string {
	return ""
}

// ReplyToMsg always returns nil since Discord doesn't have replies
func (m *Msg) ReplyToMsg() messengers.Message {
	return nil
}

// Raw returns *slackevents.EventsAPICallbackEvent
func (m *Msg) Raw() interface{} {
	return m.event
}
