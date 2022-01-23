package hangouts

import (
	"context"
	"fmt"
	"strings"

	"github.com/treeder/messengers"
	chat "google.golang.org/api/chat/v1"
)

type Msg struct {
	mess  *HangoutsMessenger
	event *chat.DeprecatedEvent
	Msg   *chat.Message
	// service *chat.Service
	cmd   string
	split []string
}

func NewMsg(ctx context.Context, mess *HangoutsMessenger, event *chat.DeprecatedEvent) *Msg {
	m := event.Message
	text := m.Text
	// replace all USER_MENTIONS with the user id/name
	stripCount := int64(0)
	for _, a := range m.Annotations {
		if a.Type == "USER_MENTION" {
			fmt.Printf("MENTION: %+v\n", a.UserMention)
			fmt.Printf("MENTION USER: %+v\n", a.UserMention.User)
			// if a.UserMention.User.Name == mess.botID() {
			// todo: should only strip this current bots names, but I can't find where to get this
			// outside talking to the bot first
			if a.UserMention.User.Type == "BOT" {
				textb := []byte(text)
				textb = append(textb[:a.StartIndex-stripCount], textb[(a.StartIndex-stripCount)+a.Length:]...)
				text = string(textb)
				stripCount += a.Length
			} else {
				// textb := []byte(text)
				// textc := append(textb[:a.StartIndex], []byte(a.UserMention.User.Name)...)
				// textc = append(textc, textb[a.Length:]...)
				// text = string(textc)
				// text = text[:(a.StartIndex-stripCount)] + a.UserMention.User.DisplayName + text[(a.StartIndex-stripCount)+a.Length:]
				// stripCount += a.Length
			}
		}
	}
	text = strings.TrimSpace(text)
	cmd, tsplit := messengers.ParseCommand(ctx, text)
	msg := &Msg{
		mess:  mess,
		event: event,
		Msg:   m,
		cmd:   cmd,
		split: tsplit,
	}
	fmt.Printf("IN MSG: %+v\n", msg)
	return msg
}

func (m *Msg) ID() string {
	return m.Msg.Name
}

func (m *Msg) TeamID() string {
	return ""
}

func (m *Msg) Command() string {
	return m.cmd
}

func (m *Msg) IsSlashCommand() bool {
	return strings.HasPrefix(m.FullText(), "/")
}

func (m *Msg) ChatID() string {
	// return strings.TrimPrefix(m.Msg.Space.Name, "spaces/")
	return m.Msg.Space.Name
}

func (m *Msg) ThreadID() string {
	return m.Msg.Thread.Name
}

func (m *Msg) FromID() string {
	return m.Msg.Sender.Name
}
func (m *Msg) FromUsername() string {
	return m.Msg.Sender.DisplayName
}
func (m *Msg) FullText() string {
	return m.Msg.Text
}
func (m *Msg) Cmd() string {
	return m.Command()
}
func (m *Msg) Split() []string {
	return m.split
}
func (m *Msg) IsPrivate() bool {
	return m.Msg.Space.Type == "DM"
}

func (m *Msg) Mention() string {
	return fmt.Sprintf("<%s>", m.Msg.Sender.Name)
}

// ReplyToMsgID always returns "" since Discord doesn't have replies
func (m *Msg) ReplyToMsgID() string {
	return ""
}

// ReplyToMsg always returns nil since Discord doesn't have replies
func (m *Msg) ReplyToMsg() messengers.Message {
	return nil
}

// func (m *Msg) SpaceID() string {
// 	return strings.TrimPrefix(m.Msg.Space.Name, "spaces/")
// }

// Raw returns *chat.DeprecatedEvent
func (m *Msg) Raw() interface{} {
	return m.Msg
}
