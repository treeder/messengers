package messengers

import (
	"context"
	"net/http"

	"github.com/treeder/messengers/models"
)

var (
	StickerMap = map[string][]string{
		"tip": []string{"https://media.giphy.com/media/wrBURfbZmqqXu/giphy-tumblr.gif", "https://media.giphy.com/media/sDcfxFDozb3bO/200w_d.gif",
			"https://media.giphy.com/media/mHcEcam5FtKQE/200w_d.gif", "https://media.giphy.com/media/l0HFkA6omUyjVYqw8/200w_d.gif",
			"https://media.giphy.com/media/uFtywzELtkFzi/200w_d.gif", "https://media.giphy.com/media/gTURHJs4e2Ies/200w_d.gif"},
		// More tip ones: https://gph.is/2Ldm07f ,
		"lucky":       []string{"https://raw.githubusercontent.com/treeder/messengers/main/assets/images/lucky1.jpg"},
		"donate":      []string{"https://media.giphy.com/media/13MWKSosOiRpe0/giphy-downsized.gif"},
		"redenvelope": []string{"https://github.com/treeder/messengers/blob/main/assets/images/red-envelope-1.png"},
	}
)

const (
	ServiceFirebase = "firebase"
	ServiceTelegram = "telegram"
	ServiceDiscord  = "discord"
	ServiceHangouts = "hangouts"
	ServiceSlack    = "slack"
	ServiceSMS      = "sms"
)

type Messenger interface {
	// Name returns the name of the messenger, eg: telegram, discord
	Name() string
	SendMsg(ctx context.Context, in IncomingMessage, text string, opts SendOpts) (Message, error)
	SendMsgMulti(ctx context.Context, in IncomingMessage, text []string, opts SendOpts) (Message, error)
	SendMsgTo(ctx context.Context, chatID, text string, opts SendOpts) (Message, error)
	SendMsgToUser(ctx context.Context, su *models.ServiceUser, text string, opts SendOpts) (Message, error)
	EditMsg(ctx context.Context, toEdit Message, text string, opts SendOpts) (Message, error)
	// should maybe deprecate below, lacks important information like Team that can be provided in the Message object
	EditMsg2(ctx context.Context, chatID, msgID, text string, opts SendOpts) (Message, error)
	SendError(ctx context.Context, in IncomingMessage, err error) (Message, error)
	MentionBot() string
	Mention(*models.ServiceUser) string
	// Returns true if the mention is the bot itself
	MentionIsMe(ctx context.Context, usernameOrID string) bool
	ExtractIDFromMention(in IncomingMessage, mention string) (string, error)
	// Format applies the appropriate formatting based on the messenger
	Format(f FormatStr, s interface{}) string
	Link(text, url string) string
	HelpMsgAddToGroup() string

	// if needed to shut down
	Close()

	ChatInfo(ctx context.Context, in IncomingMessage) (*ChatInfo, error)

	// ForOrg clones the messenger and sets the org specific oauth access token for the bot
	// ForOrg(oauthToken string) Messenger

	AddHandler(ctx context.Context, h MessageHandler)
	HandleEventHTTP(w http.ResponseWriter, r *http.Request)
}

type BaseMessenger struct {
	MessageHandlers    []MessageHandler
	messageHandlersRaw []MessageHandlerRaw
}

func (mess *BaseMessenger) AddHandler(ctx context.Context, h MessageHandler) {
	// fmt.Printf("Adding handler: %+v\n", h)
	mess.MessageHandlers = append(mess.MessageHandlers, h)
	// fmt.Printf("handler slice after: %+v\n", mess.MessageHandlers)
}

// func (mess *BaseMessenger) AddHandlerRaw(ctx context.Context, h MessageHandler) {
// 	mess.messageHandlers = append(mess.messageHandlers, h)
// }

type SendOpts map[string]interface {
	
}

// func NewSendOpts(m map[string]interface{}){
// 	return &SendOpts{Map: m}
// }
