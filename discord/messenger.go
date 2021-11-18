package discord

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/treeder/gotils/v2"
	"github.com/treeder/messengers"
	"github.com/treeder/messengers/models"
)

type DiscordMessenger struct {
	sess     *discordgo.Session
	ClientID string
	// messageHandlers []messengers.MessageHandler
	*messengers.BaseMessenger
	baseCtx context.Context
}

func (mess *DiscordMessenger) ForOrg(oauthToken string) messengers.Messenger {
	newM := *mess
	return &newM
}
func (m *DiscordMessenger) Close() {
	m.sess.Close()
}

func (m *DiscordMessenger) Name() string {
	return "discord"
}

func (m *DiscordMessenger) ChatInfo(ctx context.Context, in messengers.IncomingMessage) (*messengers.ChatInfo, error) {

	chatInfo := &messengers.ChatInfo{}
	msg := in.(*InMsg)
	chatInfo.RoomID = msg.Msg.ChannelID
	// chatInfo.RoomName = msg.Msg.
	// chatInfo.ThreadID = msg.Msg.Thread.Name
	return chatInfo, nil
}

func (m *DiscordMessenger) SendMsg(ctx context.Context, in messengers.IncomingMessage, text string, opts messengers.SendOpts) (messengers.Message, error) {
	return m.SendMsgTo(ctx, in.ChatID(), text, opts)
}

func (m *DiscordMessenger) SendMsgTo(ctx context.Context, chatID, text string, opts messengers.SendOpts) (messengers.Message, error) {
	if opts == nil {
		opts = messengers.SendOpts{}
	}
	embed := m.makeEmbed(text, opts)
	m2, err := m.sess.ChannelMessageSendEmbed(chatID, embed)
	if err != nil {
		return nil, gotils.C(ctx).Errorf("error sending message:", err)
	}
	// in2 := in.(*InMsg)
	return &InMsg{Msg: m2, sess: m.sess, embed: embed}, nil
}

func (m *DiscordMessenger) SendMsgToUser(ctx context.Context, su *models.ServiceUser, text string, opts messengers.SendOpts) (messengers.Message, error) {
	return m.SendMsgTo(ctx, su.ID, text, opts)
}

func (m *DiscordMessenger) makeEmbed(text string, opts messengers.SendOpts) *discordgo.MessageEmbed {
	sticker := ""
	if opts["sticker"] != nil {
		sticker = opts["sticker"].(string)
		sticker = messengers.ChooseRandom(messengers.StickerMap[sticker])
	}
	// m2, err := m.sess.ChannelMessageSend(chatID, text)
	// This helps: https://leovoel.github.io/embed-visualizer/
	embed := &discordgo.MessageEmbed{
		Color:       8041132,
		Description: text,
	}
	if sticker != "" {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: sticker,
		}
	}
	if opts["image"] != nil {
		image := opts["image"].(string)
		embed.Image = &discordgo.MessageEmbedImage{
			URL: image,
		}
	}
	return embed
}

func (m *DiscordMessenger) SendError(ctx context.Context, in messengers.IncomingMessage, err error) (messengers.Message, error) {
	return m.SendMsg(ctx, in, err.Error(), nil)
}

func (m *DiscordMessenger) EditMsg(ctx context.Context, in messengers.Message, text string, opts messengers.SendOpts) (messengers.Message, error) {
	if opts == nil {
		opts = messengers.SendOpts{}
	}
	in2 := in.(*InMsg)
	// m2, err := m.sess.ChannelMessageEdit(in.ChatID(), in2.Msg.ID, text)
	embed := in2.embed
	if embed == nil {
		embed = m.makeEmbed(text, opts)
	} else {
		embed.Description = text
	}
	m2, err := m.sess.ChannelMessageEditEmbed(in.ChatID(), in2.Msg.ID, embed)
	if err != nil {
		return nil, gotils.C(ctx).Errorf("error editing message:", err)
	}
	return &InMsg{Msg: m2, sess: m.sess}, nil
}

func (m *DiscordMessenger) EditMsg2(ctx context.Context, chatID, msgID, text string, opts messengers.SendOpts) (messengers.Message, error) {
	min, err := m.sess.ChannelMessage(chatID, msgID)
	if err != nil {
		return nil, err
	}

	// we're always using embeds right now in discord... so:
	embed := min.Embeds[0]
	embed.Description = text
	in2 := &InMsg{
		Msg:   min,
		embed: embed,
	}
	m2, err := m.sess.ChannelMessageEditEmbed(in2.ChatID(), in2.Msg.ID, embed)
	if err != nil {
		return nil, gotils.C(ctx).Errorf("error editing message 2:", err)
	}
	return &InMsg{Msg: m2, sess: m.sess}, nil
}

func (m *DiscordMessenger) MentionBot() string {
	return fmt.Sprintf("<@%s>", m.ClientID)
}

func (m *DiscordMessenger) Mention(su *models.ServiceUser) string {
	return fmt.Sprintf("<@%s>", su.ID)
}

func (m *DiscordMessenger) MentionIsMe(ctx context.Context, usernameOrID string) bool {
	return usernameOrID == m.ClientID
}

func (m *DiscordMessenger) ExtractIDFromMention(in messengers.IncomingMessage, mention string) (string, error) {
	if !strings.HasPrefix(mention, "<@") {
		// then not a mention
		return "", fmt.Errorf("invalid username: %v", mention)
	}
	re := regexp.MustCompile("[<@!>]")
	return re.ReplaceAllString(mention, ""), nil

}

func (m *DiscordMessenger) Format(f messengers.FormatStr, s interface{}) string {
	switch f {
	case messengers.Bold:
		return fmt.Sprintf("**%v**", s)
	case messengers.Italic:
		return fmt.Sprintf("_%v_", s)
	}
	return fmt.Sprintf("%v", s)
}

func (m *DiscordMessenger) Link(text, url string) string {
	return fmt.Sprintf("[%v](%v)", text, url)
}

func (m *DiscordMessenger) HelpMsgAddToGroup() string {
	return "🤠 Be sure to add me to your servers! Just [click this link](https://discordapp.com/api/oauth2/authorize?client_id=497421043518930964&scope=bot&permissions=313408) to add me. 🙏"
}
func (mess *DiscordMessenger) AddHandler(ctx context.Context, h messengers.MessageHandler) {
	mess.MessageHandlers = append(mess.MessageHandlers, h)
}

func (mess *DiscordMessenger) HandleEventHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "HTTP not supported with discord yet.", http.StatusNotImplemented)
}