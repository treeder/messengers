package slack

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/slack-go/slack"
	"github.com/treeder/gotils/v2"
	"github.com/treeder/messengers"
	"github.com/treeder/messengers/datastore"
	"github.com/treeder/messengers/models"
)

type SlackMessenger struct {
	*messengers.BaseMessenger
	devBotToken      string
	clientID         string
	clientSecret     string
	signingSecret    string
	oauthRedirectURI string

	// per team basis
	botOauthToken string

	db *firestore.Client
}

func (mess *SlackMessenger) ForOrg(oauthToken string) messengers.Messenger {
	newM := *mess
	newM.botOauthToken = oauthToken
	return &newM
}

func (m *SlackMessenger) Name() string {
	return "slack"
}

func (m *SlackMessenger) ChatInfo(ctx context.Context, in messengers.IncomingMessage) (*messengers.ChatInfo, error) {
	return &messengers.ChatInfo{
		TeamID:   in.TeamID(),
		RoomName: "",
		RoomID:   in.ChatID(),
		ThreadID: in.ThreadID(),
	}, nil
}

func (m *SlackMessenger) SendMsg(ctx context.Context, in messengers.IncomingMessage, text string, opts messengers.SendOpts) (messengers.Message, error) {
	if opts == nil {
		opts = messengers.SendOpts{}
	}
	in2 := in.(*Msg)
	opts["teamID"] = in2.TeamID()
	opts["threadID"] = in2.ThreadID()
	return m.SendMsgTo(ctx, in.ChatID(), text, opts)
}
func (m *SlackMessenger) SendMsgMulti(ctx context.Context, in messengers.IncomingMessage, text []string, opts messengers.SendOpts) (messengers.Message, error) {
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

func (m *SlackMessenger) SlackClient(ctx context.Context) *slack.Client {
	team := ctx.Value("team").(*models.Team)
	// fmt.Printf("SLACK CLIENT TEAM: %+v", team)
	return slack.New(team.BotAccessToken)
}

// returns the client associated with the team referenced by chatID
func (m *SlackMessenger) SlackClient2(ctx context.Context, teamID string) (*models.Team, *slack.Client, error) {
	ctx = gotils.With(ctx, "function", "SlackClient2")
	// idSplit := strings.Split(chatID, "-")
	team, err := datastore.GetTeam(ctx, m.db, messengers.ServiceSlack, teamID)
	if err != nil {
		return nil, nil, gotils.C(ctx).Errorf("error getting team, this is BAD! %v", err)
	}
	return team, slack.New(team.BotAccessToken), nil
}

func (m *SlackMessenger) SendMsgTo(ctx context.Context, chatID, text string, opts messengers.SendOpts) (messengers.Message, error) {
	// had to hack teams in...
	if opts == nil {
		opts = messengers.SendOpts{}
	}
	if opts["teamID"] == nil || opts["teamID"].(string) == "" {
		return nil, gotils.C(ctx).Errorf("Slack SendMsgTo requires teamID in opts")
	}
	teamID := opts["teamID"].(string)
	_, client, err := m.SlackClient2(ctx, teamID)
	if err != nil {
		return nil, err
	}
	// idSplit := strings.Split(in.ChatID(), "-")
	// returns response channel and response timestamp (timestamp is the msgID / threadID)
	moptions := []slack.MsgOption{
		slack.MsgOptionText(fmt.Sprintf("%v", text), false),
	}
	if opts["threadID"] != nil {
		threadID := opts["threadID"].(string)
		if threadID != "" {
			moptions = append(moptions, slack.MsgOptionTS(threadID))
		}
	}
	if opts["reply_broadcast"] != nil && opts["reply_broadcast"].(bool) {
		moptions = append(moptions, slack.MsgOptionBroadcast())
	}
	fmt.Println("TEAMID:", teamID, "CHANNELID:", chatID)
	channelID, msgID, err := client.PostMessageContext(ctx, chatID, moptions...)
	if err != nil {
		return nil, gotils.C(ctx).Errorf("error on PostMessage in SendMsg", err)
	}
	// fmt.Println("SENT MSG: channel: ", channelID, "timestamp:", msgID)
	// in2 := in.(*Msg)
	in2 := &Msg{ctx: ctx}
	in2.chatID = chatID
	in2.postedID = msgID
	in2.postedChannelID = channelID
	return in2, nil
}

func (m *SlackMessenger) SendMsgToUser(ctx context.Context, su *models.ServiceUser, text string, opts messengers.SendOpts) (messengers.Message, error) {
	return m.SendMsgTo(ctx, su.DMChatID, text, opts)
}

func (m *SlackMessenger) SendError(ctx context.Context, in messengers.IncomingMessage, err error) (messengers.Message, error) {
	return m.SendMsg(ctx, in, err.Error(), nil)
}

func (m *SlackMessenger) EditMsg(ctx context.Context, in messengers.Message, text string, opts messengers.SendOpts) (messengers.Message, error) {
	if opts == nil {
		opts = messengers.SendOpts{}
	}
	// in2 := in.(*Msg)
	return m.EditMsg2(ctx, in.ChatID(), in.ID(), text, opts)
}

func (m *SlackMessenger) EditMsg2(ctx context.Context, chatID, msgID, text string, opts messengers.SendOpts) (messengers.Message, error) {
	// fmt.Println("EDITING MSG: channel: ", chatID, "timestamp:", msgID)
	_, client, err := m.SlackClient2(ctx, chatID)
	if err != nil {
		return nil, err
	}
	idSplit := strings.Split(chatID, "-")
	channelID, msgID, _, err := client.UpdateMessageContext(ctx, idSplit[1], msgID, slack.MsgOptionText(fmt.Sprintf("%v", text), false))
	if err != nil {
		return nil, gotils.C(ctx).Errorf("error on UpdateMessage in EditMsg2", err)
	}
	in2 := &Msg{ctx: ctx}
	in2.chatID = chatID
	in2.postedID = msgID
	in2.postedChannelID = channelID
	return in2, nil
}

func (m *SlackMessenger) MentionBot() string {
	return fmt.Sprintf("<@%s>", "somebot")
}

func (m *SlackMessenger) Mention(su *models.ServiceUser) string {
	id := su.ID
	split := strings.Split(su.ID, "-")
	if len(split) > 1 {
		id = split[1]
	}
	return fmt.Sprintf("<@%s>", id)
}

// bod IDs are per team I think
// func (m *SlackMessenger) botID() string {
// 	if m.infra.IsDev() {
// 		return (devBotID)
// 	}
// 	return "NOTHING YET"
// }

func (m *SlackMessenger) MentionIsMe(ctx context.Context, usernameOrID string) bool {
	team := ctx.Value("team").(*models.Team)
	return usernameOrID == team.BotUserID // this is unique per team...
}

func (m *SlackMessenger) ExtractIDFromMention(in messengers.IncomingMessage, mention string) (string, error) {
	if !strings.HasPrefix(mention, "<@") {
		// then not a mention
		return "", fmt.Errorf("invalid username: %v", mention)
	}
	re := regexp.MustCompile("[<@!>]")
	id := re.ReplaceAllString(mention, "")
	in2 := in.(*Msg)
	return in2.TeamID() + "-" + id, nil

}

func (m *SlackMessenger) Format(f messengers.FormatStr, s interface{}) string {
	switch f {
	case messengers.Bold:
		return fmt.Sprintf("*%v*", s)
	case messengers.Italic:
		return fmt.Sprintf("_%v_", s)
	}
	return fmt.Sprintf("%v", s)
}

func (m *SlackMessenger) Link(text, url string) string {
	return fmt.Sprintf("<%v|%v>", url, text)
}

func (m *SlackMessenger) HelpMsgAddToGroup() string {
	return "🤠 Be sure to add me to your Slack channels! 🙏"
}

func (m *SlackMessenger) Close() {}
