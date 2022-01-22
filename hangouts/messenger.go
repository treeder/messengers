package hangouts

import (
	"context"
	"fmt"
	"strings"

	"github.com/treeder/gotils/v2"
	"github.com/treeder/messengers"
	"github.com/treeder/messengers/models"
	chat "google.golang.org/api/chat/v1"
)

type HangoutsMessenger struct {
	_botID string
	hchat  *chat.Service
	*messengers.BaseMessenger
}

func (mess *HangoutsMessenger) ForOrg(oauthToken string) messengers.Messenger {
	newM := *mess
	return &newM
}

func (m *HangoutsMessenger) Name() string {
	return "hangouts"
}

func (m *HangoutsMessenger) ChatInfo(ctx context.Context, in messengers.IncomingMessage) (*messengers.ChatInfo, error) {
	// hangouts := GetMessenger(ServiceHangouts)
	// hangouts := hangoutsTmp.(*hangouts.HangoutsMessenger)
	api := m.hchat
	// THIS IS HOW YOU CAN FIND THE SPACE ID TO DO THESE CHATS IN
	// COULD ALSO HAVE A @bot info command that returns current space ID
	listresp, err := api.Spaces.List().Do()
	if err != nil {
		return nil, gotils.C(ctx).Errorf("error getting spaces list: %v", (err))
		// return err
	}
	for _, s := range listresp.Spaces {
		fmt.Printf("%+v\n", s)
	}
	chatInfo := &messengers.ChatInfo{}
	msg := in.(*Msg)
	chatInfo.RoomID = msg.Msg.Space.Name
	chatInfo.RoomName = msg.Msg.Space.DisplayName
	chatInfo.ThreadID = msg.Msg.Thread.Name
	return chatInfo, nil
}

func (m *HangoutsMessenger) SendMsg(ctx context.Context, in messengers.IncomingMessage, text string, opts messengers.SendOpts) (messengers.Message, error) {
	// return m.SendMsgTo(ctx, in.ChatID(), text, opts)
	if opts == nil {
		opts = messengers.SendOpts{}
	}
	// in2 := in.(*Msg)
	fmt.Printf("in: %+v\nhchat: %+v\n", in, m.hchat)
	res, err := m.hchat.Spaces.Messages.Create(in.ChatID(), &chat.Message{
		Thread: &chat.Thread{
			Name: in.ThreadID(),
		},
		Text: text,
	}).Do()
	if err != nil {

		return nil, gotils.C(ctx).Errorf("error sending message: %v", err)
	}
	// fmt.Printf("MESSAGE RESULT: %+v\n", *res)

	return &Msg{Msg: res}, nil
}

func (m *HangoutsMessenger) SendMsgMulti(ctx context.Context, in messengers.IncomingMessage, text []string, opts messengers.SendOpts) (messengers.Message, error) {
	var msg messengers.Message
	var err error
	for _, s := range text {
		msg, err = m.SendMsg(ctx, in, s, opts)
		if err != nil {
			return msg, err
		}
	}
	return msg, nil
}

func (m *HangoutsMessenger) SendMsgTo(ctx context.Context, chatID, text string, opts messengers.SendOpts) (messengers.Message, error) {
	if opts == nil {
		opts = messengers.SendOpts{}
	}

	call := m.hchat.Spaces.Messages.Create(chatID, &chat.Message{
		// Thread: &chat.Thread{
		// 	Name: chatID,
		// },
		Text: text,
	})
	if opts["threadKey"] != nil {
		threadKey := opts["threadKey"].(string)
		call = call.ThreadKey(threadKey)
	}

	res, err := call.Do()
	if err != nil {
		return nil, gotils.C(ctx).Errorf("error sending message to: %v", err)
	}
	// fmt.Printf("MESSAGE RESULT: %+v\n", *res)

	return &Msg{Msg: res}, nil
}

func (m *HangoutsMessenger) SendMsgToUser(ctx context.Context, su *models.ServiceUser, text string, opts messengers.SendOpts) (messengers.Message, error) {
	return m.SendMsgTo(ctx, su.DMChatID, text, opts)
}

func (m *HangoutsMessenger) SendError(ctx context.Context, in messengers.IncomingMessage, err error) (messengers.Message, error) {
	return m.SendMsg(ctx, in, err.Error(), nil)
}

func (m *HangoutsMessenger) EditMsg(ctx context.Context, in messengers.Message, text string, opts messengers.SendOpts) (messengers.Message, error) {
	if opts == nil {
		opts = messengers.SendOpts{}
	}
	// in2 := in.(*Msg)
	return m.EditMsg2(ctx, in.ChatID(), in.ID(), text, opts)

}

func (m *HangoutsMessenger) EditMsg2(ctx context.Context, chatID, msgID, text string, opts messengers.SendOpts) (messengers.Message, error) {

	// currently only text/cards in fieldmask https://developers.google.com/hangouts/chat/reference/rest/v1/spaces.messages/update
	res, err := m.hchat.Spaces.Messages.Update(msgID, &chat.Message{
		// Thread: &chat.Thread{
		// 	Name: chatID,
		// },
		Text: text,
	}).UpdateMask("text").Do()
	if err != nil {
		return nil, gotils.C(ctx).Errorf("error editing message: %v", err)
	}
	// fmt.Printf("MESSAGE RESULT: %+v\n", *res)
	return &Msg{Msg: res}, nil
}

func (m *HangoutsMessenger) MentionBot() string {
	return fmt.Sprintf("<%s>", m.botID())
}
func (m *HangoutsMessenger) Mention(su *models.ServiceUser) string {
	return fmt.Sprintf("<%s>", su.ID)
}

func (m *HangoutsMessenger) botID() string {
	return m._botID
}

func (m *HangoutsMessenger) MentionIsMe(ctx context.Context, usernameOrID string) bool {
	return usernameOrID == m.botID()
}

func (m *HangoutsMessenger) ExtractIDFromMention(in messengers.IncomingMessage, mention string) (string, error) {
	if !strings.HasPrefix(mention, "users/") {
		return "", fmt.Errorf("invalid username: %v", mention)
	}
	// hangoutsUserID := strings.TrimPrefix(mention, "users/")
	// return hangoutsUserID, nil
	return mention, nil
}

func (m *HangoutsMessenger) Format(f messengers.FormatStr, s interface{}) string {
	switch f {
	case messengers.Bold:
		return fmt.Sprintf("*%v*", s)
	case messengers.Italic:
		return fmt.Sprintf("_%v_", s)
	}
	return fmt.Sprintf("%v", s)
}

func (m *HangoutsMessenger) Link(text, url string) string {
	return fmt.Sprintf("<%v|%v>", url, text)
}

func (m *HangoutsMessenger) HelpMsgAddToGroup() string {
	return "🤠 Be sure to add me to your Hangout rooms! 🙏"
}

func (m *HangoutsMessenger) Close() {}
