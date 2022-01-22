package sms

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/sfreiberg/gotwilio"
	"github.com/treeder/gotils/v2"
	"github.com/treeder/messengers"
	"github.com/treeder/messengers/models"
)

func New(ctx context.Context, db *firestore.Client, twilioID, twilioToken string) (*SMSMessenger, error) {
	return &SMSMessenger{
		BaseMessenger: &messengers.BaseMessenger{},
		client:        gotwilio.NewTwilioClient(twilioID, twilioToken),
	}, nil
}

type SMSMessenger struct {
	*messengers.BaseMessenger
	client *gotwilio.Twilio
}

func (mess *SMSMessenger) Client() *gotwilio.Twilio {
	return mess.client
}

func (mess *SMSMessenger) LookupNumber(num string) (gotwilio.Lookup, error) {
	return mess.client.LookupNoCarrier(num)
}

func (mess *SMSMessenger) ForOrg(oauthToken string) messengers.Messenger {
	newM := *mess
	return &newM
}

func (mess *SMSMessenger) Name() string {
	return "sms"
}
func (mess *SMSMessenger) SendMsg(ctx context.Context, in messengers.IncomingMessage, text string, opts messengers.SendOpts) (messengers.Message, error) {
	return mess.SendMsgTo(ctx, in.FromID(), text, messengers.SendOpts{
		"botNumber": in.ChatID(),
	})
}

func (m *SMSMessenger) SendMsgMulti(ctx context.Context, in messengers.IncomingMessage, text []string, opts messengers.SendOpts) (messengers.Message, error) {
	var msg messengers.Message
	var err error
	for _, s := range text {
		msg, err = m.SendMsgTo(ctx, in.ChatID(), s, opts)
		if err != nil {
			return msg, err
		}
	}
	return msg, nil
}

// SendMsgTo requires a from phone number in the options. SMS works differently than the rest, so had to add
// that.
func (mess *SMSMessenger) SendMsgTo(ctx context.Context, chatID, text string, opts messengers.SendOpts) (messengers.Message, error) {
	fmt.Printf("TWILIO sending msg to %v, from %v: %v\n", chatID, opts["botNumber"], text)
	resp, exception, err := mess.client.SendSMS(opts["botNumber"].(string), chatID, text, "", "")
	fmt.Printf("TWILIO RESP: %+v\n", resp)
	fmt.Printf("TWILIO EXCEPTION: %+v\n", exception)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
func (mess *SMSMessenger) SendMsgToUser(ctx context.Context, su *models.ServiceUser, text string, opts messengers.SendOpts) (messengers.Message, error) {
	return nil, nil
}
func (mess *SMSMessenger) EditMsg(ctx context.Context, toEdit messengers.Message, text string, opts messengers.SendOpts) (messengers.Message, error) {
	return nil, nil
}
func (mess *SMSMessenger) EditMsg2(ctx context.Context, chatID, msgID, text string, opts messengers.SendOpts) (messengers.Message, error) {
	return nil, nil
}
func (mess *SMSMessenger) SendError(ctx context.Context, in messengers.IncomingMessage, err error) (messengers.Message, error) {
	return mess.SendMsg(ctx, in, err.Error(), nil)
}
func (mess *SMSMessenger) MentionBot() string {
	return ""
}
func (mess *SMSMessenger) Mention(*models.ServiceUser) string {
	return ""
}
func (mess *SMSMessenger) MentionIsMe(ctx context.Context, usernameOrID string) bool {
	return false
}
func (mess *SMSMessenger) ExtractIDFromMention(in messengers.IncomingMessage, mention string) (string, error) {
	return "", nil
}
func (mess *SMSMessenger) Format(f messengers.FormatStr, s interface{}) string {
	return ""
}
func (mess *SMSMessenger) Link(text, url string) string {
	return ""
}
func (mess *SMSMessenger) HelpMsgAddToGroup() string {
	return ""
}
func (mess *SMSMessenger) Close() {}
func (mess *SMSMessenger) ChatInfo(ctx context.Context, in messengers.IncomingMessage) (*messengers.ChatInfo, error) {
	return nil, nil
}

func (mess *SMSMessenger) HandleEventHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = gotils.With(ctx, "messenger", "sms")

	m := &gotwilio.SMSWebhook{}
	err := r.ParseForm()
	if err != nil {
		gotils.LogBeta(ctx, "error", "error on ParseForm: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = gotwilio.DecodeWebhook(r.Form, m)
	if err != nil {
		gotils.LogBeta(ctx, "error", "error on sms.DecodeWebhook", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Printf("WEBHOOK: %+v\n", m)

	msg := &Message{
		raw: m,
	}
	text := strings.TrimSpace(m.Body)
	msg.cmd, msg.split = messengers.ParseCommand(ctx, text)
	for _, h := range mess.MessageHandlers {
		// fmt.Println("handling message", h)
		h.HandleMessage(ctx, mess, msg)
	}
	w.WriteHeader(http.StatusOK)
}
