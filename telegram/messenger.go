package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/treeder/gotils/v2"
	"github.com/treeder/messengers"
	"github.com/treeder/messengers/models"
	"github.com/treeder/messengers/ui"
)

const (
	StickerBullTakeMyMoney      = "CAADAgADMwMAApzW5wqo8mwODr2YhwI"
	StickerBullMoneyParty       = "CAADAgADLwMAApzW5wqe8VvH7rRlRwI"
	StickerBullTipFlip          = "CAADAgADMAMAApzW5wrDWJktJgLrQQI"
	StickerBullThummpsUp        = "CAADAgADEwMAApzW5wr-7CmoPjlOYwI"
	StickerBullJustHODL         = "CAADAgADHQMAApzW5wrQArgAAaBpjJsC"
	StickerMonoPleasureDoingBiz = "CAADAgADgwEAAgeGFQdsHwABu9sMwMgC"
	StickerMonoLuckyMan         = "CAADAgADiwEAAgeGFQeIXf49-peWNgI"
	StickerBelfortThumbsUp      = "CAADAgAD2AEAAzigCmusViuaetMSAg"
	StickerAirdropMonster       = "CAADAgADwwADW__yCgk0jkn5y6AaAg"
)

var (
	stickerMap = map[string]string{
		"tip":         StickerBullTipFlip,
		"lucky":       StickerMonoLuckyMan,
		"donate":      StickerBullMoneyParty,
		"redenvelope": "http://icons.iconarchive.com/icons/gcds/chinese-new-year/256/red-envelope-icon.png",
	}
)

type TelegramMessenger struct {
	*messengers.BaseMessenger
	tg                *tgbotapi.BotAPI
	apiKey            string
	rsaKeyForPassport []byte

	db *firestore.Client
}

// Client returns the underlying telegram client
func (m *TelegramMessenger) Client() *tgbotapi.BotAPI {
	return m.tg
}

func (m *TelegramMessenger) Name() string {
	return "telegram"
}

func (m *TelegramMessenger) ChatInfo(ctx context.Context, in messengers.IncomingMessage) (*messengers.ChatInfo, error) {

	chatInfo := &messengers.ChatInfo{}
	msg := in.(*InMsg)
	chatInfo.RoomID = strconv.FormatInt(msg.Msg.Chat.ID, 10)
	chatInfo.RoomName = msg.Msg.Chat.Title
	// chatInfo.ThreadID = msg.Msg.Thread.Name
	return chatInfo, nil
}

func (mess *TelegramMessenger) SendMsg(ctx context.Context, in messengers.IncomingMessage, text string, opts messengers.SendOpts) (messengers.Message, error) {
	return mess.SendMsgTo(ctx, in.ChatID(), text, opts)
}

func (mess *TelegramMessenger) SendMsgTo(ctx context.Context, chatID, text string, opts messengers.SendOpts) (messengers.Message, error) {
	if opts == nil {
		opts = messengers.SendOpts{}
	}
	ctx = gotils.With(ctx, "chat_id", chatID)
	mto, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {

		return nil, gotils.C(ctx).Errorf("error converting chat id to int64: %v", (err))
	}
	m2, err := mess.sendMsgTo2(ctx, mto, text, opts)
	if err != nil {
		return nil, err
	}
	return &InMsg{Msg: m2}, nil
}

func (mess *TelegramMessenger) SendMsgToUser(ctx context.Context, su *models.ServiceUser, text string, opts messengers.SendOpts) (messengers.Message, error) {
	return mess.SendMsgTo(ctx, su.ID, text, opts)
}

// sendMsgTo2 returns the Message as well
func (mess *TelegramMessenger) sendMsgTo2(ctx context.Context, chatID int64, text string, opts messengers.SendOpts) (*tgbotapi.Message, error) {
	var m *tgbotapi.Message
	// This will use the text as an image caption if "image" value is present and will NOT send another regular message.
	if opts["image"] != nil {
		image := opts["image"].(string)
		// smsg := tgbotapi.NewStickerShare(chatID, image)
		smsg := tgbotapi.NewPhotoShare(chatID, image+"?"+messengers.GenToken())
		// smsg.MimeType = "image/png"
		// if opts["caption"] != nil {
		// caption := opts["caption"].(string)
		smsg.Caption = text
		smsg.ParseMode = tgbotapi.ModeMarkdown
		// }
		m, err := mess.tg.Send(smsg)
		if err != nil {
			gotils.LogBeta(ctx, "error", "Error sending image %v to telegram: %v", image, (err))
			// this could be important for collectibles
			merr := tgbotapi.NewMessage(chatID, fmt.Sprintf("Telegram not happy right now, but you can %v", Link("view it here", image)))
			merr.ParseMode = tgbotapi.ModeMarkdown
			merr.DisableWebPagePreview = true
			_, err = mess.tg.Send(merr)
			if err != nil {
				gotils.LogBeta(ctx, "error", "Error sending follow up message to telegram", (err))
			}
		}
		// fmt.Printf("IMAGE RESPONSE: %+v)
		return &m, nil
	}
	if text != "" {
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.DisableWebPagePreview = true
		if opts != nil {
			if opts["buttons"] != nil {
				bslice := opts["buttons"].([]*ui.Button)
				buttons := []tgbotapi.InlineKeyboardButton{}
				for _, b := range bslice {
					buttons = append(buttons, tgbotapi.InlineKeyboardButton{Text: b.Text, CallbackData: &b.Data})
				}
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons)

			}
			if opts["reply_buttons"] != nil {
				bslice := opts["reply_buttons"].([]*ui.Button)
				buttons := []tgbotapi.KeyboardButton{}
				for _, b := range bslice {
					buttons = append(buttons, tgbotapi.KeyboardButton{Text: b.Text}) //, CallbackData: &b.Data})
				}
				rm := tgbotapi.NewReplyKeyboard(buttons)
				rm.OneTimeKeyboard = true
				// selective and I'm pretty sure one time does NOT work:
				// https://github.com/telegraf/telegraf/issues/344
				// https://github.com/yagop/node-telegram-bot-api/issues/541
				rm.Selective = true
				msg.ReplyMarkup = rm
			}
			if opts["reply_to_id"] != nil {
				replyToID, _ := strconv.Atoi(opts["reply_to_id"].(string))
				msg.ReplyToMessageID = replyToID
			}
		}
		m2, err := mess.tg.Send(msg)
		if err != nil {
			return nil, gotils.C(ctx).Errorf("Error sending message to telegram: %v", (err))
		}
		m = &m2
	}
	sticker := ""
	if opts["sticker"] != nil {
		sticker = opts["sticker"].(string)
		sticker = stickerMap[sticker]
	}
	if sticker != "" {
		var smsg tgbotapi.Chattable
		if strings.HasPrefix(sticker, "http") {
			smsg = tgbotapi.NewPhotoShare(chatID, sticker)
		} else {
			smsg = tgbotapi.NewStickerShare(chatID, sticker)
		}
		_, err := mess.tg.Send(smsg)
		if err != nil {
			gotils.LogBeta(ctx, "error", "Error sending sticker to telegram: %v", (err))
			// return err , not returning this one since it's not important
		}
	}
	return m, nil
}

func (mess *TelegramMessenger) EditMsg(ctx context.Context, in messengers.Message, text string, opts messengers.SendOpts) (messengers.Message, error) {
	m2 := in.(*InMsg)
	m3, err := mess.editMsg(ctx, m2.Msg, text)
	if err != nil {
		return nil, err
	}
	return &InMsg{Msg: m3}, nil
}

func (mess *TelegramMessenger) EditMsg2(ctx context.Context, chatID, msgID, text string, opts messengers.SendOpts) (messengers.Message, error) {
	msgIDInt, _ := strconv.Atoi(msgID)
	m2 := &tgbotapi.Message{
		MessageID: msgIDInt,
	}
	chatIDInt, _ := strconv.ParseInt(chatID, 10, 64)
	m2.Chat = &tgbotapi.Chat{
		ID: chatIDInt,
	}
	m3, err := mess.editMsg(ctx, m2, text)
	if err != nil {
		return nil, err
	}
	return &InMsg{Msg: m3}, nil
}

func (mess *TelegramMessenger) SendError(ctx context.Context, in messengers.IncomingMessage, err error) (messengers.Message, error) {
	m2 := in.(*InMsg)
	m3, err := mess.sendError2(ctx, m2.Msg, err)
	if err != nil {
		return nil, err
	}
	return &InMsg{Msg: m3}, nil
}
func (mess *TelegramMessenger) SendStickerTo(ctx context.Context, chatID, sticker string) (messengers.Message, error) {
	// m2 := in.(*InMsg)
	return mess.SendMsgTo(ctx, chatID, "", messengers.SendOpts{"sticker": sticker})
}

func (mess *TelegramMessenger) MentionBot() string {
	return fmt.Sprintf("@%v", mess.tg.Self.UserName)
}
func (mess *TelegramMessenger) Mention(su *models.ServiceUser) string {
	return messengers.MarkdownEscape(fmt.Sprintf("@%v", su.Username))
}

func (mess *TelegramMessenger) MentionIsMe(ctx context.Context, usernameOrID string) bool {
	return usernameOrID == mess.tg.Self.UserName
}

func (m *TelegramMessenger) ExtractIDFromMention(in messengers.IncomingMessage, mention string) (string, error) {
	if !strings.HasPrefix(mention, "@") {
		// then not a mention
		return "", fmt.Errorf("invalid username: %v", mention)
	}
	// then it's a username
	uNoAt := mention[1:]
	if len(uNoAt) == 0 {
		return "", fmt.Errorf("invalid username: %v", mention)
	}
	return uNoAt, nil
}

func (m *TelegramMessenger) Format(f messengers.FormatStr, s interface{}) string {
	switch f {
	case messengers.Bold:
		return fmt.Sprintf("*%v*", s)
	case messengers.Italic:
		return fmt.Sprintf("_%v_", s)
	}
	return fmt.Sprintf("%v", s)
}

func (m *TelegramMessenger) Link(text, url string) string {
	return Link(text, url)
}

func Link(text, url string) string {
	return fmt.Sprintf("[%v](%v)", text, url)
}

func (m *TelegramMessenger) HelpMsgAddToGroup() string {
	username := "add me to your group"
	bu, err := m.tg.GetMe()
	if err != nil {
		// log it, but try to continue
		gotils.L(context.Background()).Error().Printf("error on GetMe(): %v", err)
		username = fmt.Sprintf("type @%v", bu.UserName)

	}
	return fmt.Sprintf("🤠 Be sure to add me to your groups! Click on group name, click Add Meber then %v  🙏", username)
}

func (m *TelegramMessenger) Close() {}

func (mess *TelegramMessenger) editMsg(ctx context.Context, msgToUpdate *tgbotapi.Message, text string) (*tgbotapi.Message, error) {
	m := tgbotapi.NewEditMessageText(msgToUpdate.Chat.ID, msgToUpdate.MessageID, text)
	m.ParseMode = tgbotapi.ModeMarkdown
	m.DisableWebPagePreview = true
	mret, err := mess.tg.Send(m)
	if err != nil {
		return nil, gotils.C(ctx).Errorf("Error editing message on telegram: %v", (err))
	}
	return &mret, nil
}

func (mess *TelegramMessenger) editCaption(ctx context.Context, msgToUpdate tgbotapi.Message, text string) (*tgbotapi.Message, error) {
	m := tgbotapi.NewEditMessageCaption(msgToUpdate.Chat.ID, msgToUpdate.MessageID, text)
	m.ParseMode = tgbotapi.ModeMarkdown
	// m.DisableWebPagePreview = true
	mret, err := mess.tg.Send(m)
	if err != nil {
		return nil, gotils.C(ctx).Errorf("Error editing caption on telegram: %v", (err))
	}
	return &mret, nil
}

func (mess *TelegramMessenger) sendPhoto(ctx context.Context, chatID int64, photoURL, text string) (*tgbotapi.Message, error) {
	m := tgbotapi.NewPhotoShare(chatID, photoURL)
	m.ParseMode = tgbotapi.ModeMarkdown
	// m.DisableWebPagePreview = true
	m.Caption = text
	mret, err := mess.tg.Send(m)
	if err != nil {
		return nil, gotils.C(ctx).Errorf("Error sending photo message on telegram: %v", (err))
	}
	return &mret, nil
}

func (mess *TelegramMessenger) sendError2(ctx context.Context, msgIn *tgbotapi.Message, err error) (*tgbotapi.Message, error) {
	// TODO: make a UserError and don't show the contact support thing for that type - serr, ok := err.(*model.ModelMissingError)
	// for non UserError's show the link to contact support
	msgI := &InMsg{Msg: msgIn}
	m := fmt.Sprintf("%v%v %v", messengers.SprintUsername(msgI), err.Error(), "")
	msg := tgbotapi.NewMessage(msgIn.Chat.ID, m)
	// msg.ParseMode = tgbotapi.ModeMarkdown
	msg.DisableWebPagePreview = true
	msg.ReplyToMessageID = msgIn.MessageID // like a normal reply response
	m2, err := mess.tg.Send(msg)
	if err != nil {
		return nil, gotils.C(ctx).Errorf("Error sending error message to telegram: %v", (err))
	}
	return &m2, nil
}
