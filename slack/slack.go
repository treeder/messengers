package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/treeder/gotils/v2"
	"github.com/treeder/messengers"
	"github.com/treeder/messengers/datastore"
	"github.com/treeder/messengers/models"
)

// New ...
// SLACK_CLIENT_ID
// SLACK_CLIENT_SECRET
// SLACK_SIGNING_SECRET
// # Below are on a per team basis
// SLACK_BOT_TOKEN - pass in for the testing gsuite bot
func New(ctx context.Context, db *firestore.Client, clientID, clientSecret, signingSecret, botToken, oauthRedirectURI string) (*SlackMessenger, error) {
	if clientID == "-" {
		return nil, nil
	}
	if clientSecret == "-" {
		gotils.LogBeta(ctx, "warn", "SLACK_CLIENT_SECRET not set, won't be starting slack support.")
		return nil, nil
	}
	if signingSecret == "-" {
		gotils.LogBeta(ctx, "warn", "SLACK_SIGNING_SECRET not set, won't be starting slack support.")
		return nil, nil
	}
	// this token is for the testing Gsuite team
	if botToken == "-" {
		gotils.LogBeta(ctx, "warn", "SLACK_BOT_TOKEN not set, won't be starting slack support.")
		return nil, nil
	}

	ctx = gotils.With(ctx, "messenger", "slack")

	mess := &SlackMessenger{
		BaseMessenger:    &messengers.BaseMessenger{},
		db:               db,
		devBotToken:      botToken,
		clientID:         clientID,
		clientSecret:     clientSecret,
		signingSecret:    signingSecret,
		oauthRedirectURI: oauthRedirectURI,
	}
	gotils.LogBeta(ctx, "info", "Starting Slack bot")
	// go startReceiving(ctx, pubsubClient, hchat)
	return mess, nil
}

// VerifyRequest verifies a request coming from Slack
func (mess *SlackMessenger) VerifyRequest(req *http.Request) error {
	secretVerifier, err := slack.NewSecretsVerifier(req.Header, mess.signingSecret)
	if err != nil {
		return errors.Wrap(err, "NewSecretsVerifier failed")
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return errors.Wrap(err, "ReadAll failed")
	}

	// we need to reset the body to avoid unexpected side effects
	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	_, err = secretVerifier.Write(body)
	if err != nil {
		return errors.Wrap(err, "Ensure failed")
	}

	err = secretVerifier.Ensure()
	if err != nil {
		return errors.Wrap(err, "Ensure failed")
	}

	return nil
}

// func (mess *SlackMessenger) AddHandler(ctx context.Context, h messengers.MessageHandler) {
// 	mess.messageHandlers = append(mess.messageHandlers, h)
// }

func (mess *SlackMessenger) HandleEventHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = gotils.With(ctx, "messenger", "slack")
	err := mess.VerifyRequest(r)
	if err != nil {
		gotils.LogBeta(ctx, "error", "error on VerifyRequest: %v", (err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()
	event, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken()) //, slackevents.OptionVerifyToken(&slackevents.TokenComparator{VerificationToken: "TOKEN"}))
	if err != nil {
		gotils.LogBeta(ctx, "error", "error on slackevents.ParseEvent", (err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// need to respond in 3 seconds...
	switch event.Type {
	case slackevents.URLVerification:
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal([]byte(body), &r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text")
		w.Write([]byte(r.Challenge))
	case slackevents.CallbackEvent:
		cbEvent := event.Data.(*slackevents.EventsAPICallbackEvent)
		go mess.handleEventAsync(ctx, event, cbEvent)
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (mess *SlackMessenger) handleEventAsync(ctx context.Context, event slackevents.EventsAPIEvent, cbEvent *slackevents.EventsAPICallbackEvent) {
	ctx = gotils.With(ctx, "team_id", event.TeamID)
	ctx = gotils.With(ctx, "api_app_id", cbEvent.APIAppID)
	var team *models.Team
	var err error
	// if cbEvent.APIAppID == devAppID && event.TeamID == devTeamID {
	// 	team = &models.Team{ID: devTeamID, Name: "gopath", BotAccessToken: mess.devBotToken, BotUserID: devBotID}
	// } else {
	team, err = datastore.GetTeam(ctx, mess.db, messengers.ServiceSlack, event.TeamID)
	if err != nil {
		gotils.LogBeta(ctx, "error", "error getting team, this is BAD! %v", (err))
		// mess.infra.Errors.Report(errorreporting.Entry{
		// 	Error: fmt.Errorf("error getting team, this is BAD! %v", err),
		// 	User:  event.TeamID,
		// })
		return
	}
	// }
	// this is a hack to get this later on in SendMsgTo and things... didn't think about teams in the beginning
	// this might be able to be removed now that I'm prefixing TEAM- to all IDs.
	ctx = context.WithValue(ctx, "team", team)
	msg, rmsg, err := mess.handleEvent2(ctx, team, event, cbEvent)
	if err != nil {
		client := mess.SlackClient(ctx)
		_, _, err2 := client.PostMessage(msg.inner.Channel(), slack.MsgOptionText(fmt.Sprintf("ERROR: %v", err), false))
		if err2 != nil {
			gotils.LogBeta(ctx, "error", "error on PostMessage ERROR", err)
		}
	}
	if rmsg != nil {
		client := mess.SlackClient(ctx)
		_, _, err2 := client.PostMessage(msg.inner.Channel(), slack.MsgOptionText(fmt.Sprintf("%v", rmsg.Text), false))
		if err2 != nil {
			gotils.LogBeta(ctx, "error", "error on PostMessage", err)
		}
	}
}
func (mess *SlackMessenger) handleEvent2(ctx context.Context, team *models.Team, event slackevents.EventsAPIEvent, cbEvent *slackevents.EventsAPICallbackEvent) (*Msg, *messengers.ResponseMsg, error) {
	fmt.Println("handleEvent2")
	// holy shit nlopes/slack lib is shitty
	innerEvent := event.InnerEvent
	switch ev := innerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		// mess.client.PostMessage(ev.Channel, slack.MsgOptionText("Yes, hello.", false))
		return mess.respond(ctx, cbEvent, innerEvent)
	case *slackevents.MessageEvent:
		fmt.Println("is MessageEvent")
		if ev.SubType == "bot_message" {
			// skip messages from bots.
			break
		}
		if ev.BotID != "" {
			// this is probably better than bot_message, since "thread_broadcast" overides the bot_message subtype
			// https://api.slack.com/events/message/thread_broadcast
			break
		}
		// if ev.BotID == team.BotUserID { // this messes up if team.BotUserID is not ""
		// 	break
		// }
		// if cbEvent.APIAppID == devAppID || cbEvent.APIAppID == prodAppID {
		// 	// app id is global, bot ID is per team, so we're checking app id
		// 	// https://api.slack.com/methods/bots.info
		// 	break
		// }
		// only reply to random message if this is a direct message
		if ev.ChannelType == "im" {
			// mess.client.PostMessage(ev.Channel, slack.MsgOptionText("Yes, hello 2.", false))
			return mess.respond(ctx, cbEvent, innerEvent)
		}
		// or if it's in a thread we're interested in
		fmt.Println("TIMESTAMP:", ev.ThreadTimeStamp)
		if ev.ThreadTimeStamp != "" {
			return mess.respond(ctx, cbEvent, innerEvent)
		}
		// we could just let all of them pass through...?
	}
	return nil, nil, nil
}

func (mess *SlackMessenger) respond(ctx context.Context, event *slackevents.EventsAPICallbackEvent, inner slackevents.EventsAPIInnerEvent) (*Msg, *messengers.ResponseMsg, error) {
	fmt.Println("respond func")
	msg := NewMsg(ctx, mess, event, inner)
	if msg.cmd == "" {
		return msg, &messengers.ResponseMsg{Text: "How can I help you?"}, nil
	}
	var rmsg *messengers.ResponseMsg
	for _, h := range mess.MessageHandlers {
		fmt.Println("handler: ", h)
		h.HandleMessage(ctx, mess, msg)
		// r, err := h.HandleMessage(ctx, mess, msg)
		// if r != nil {
		// 	rmsg = r
		// }
		// if err != nil {
		// 	gotils.LogBeta(ctx, "error", "error on ProcessMessage2", err)
		// }

	}
	return msg, rmsg, nil
}

/*
OauthHandler ...
https://api.slack.com/docs/oauth

From oauth request:

```json
{
    "access_token": "xoxp-XXXXXXXX-XXXXXXXX-XXXXX",
    "scope": "incoming-webhook,commands,bot",
    "team_name": "Team Installing Your Hook",
    "team_id": "XXXXXXXXXX",
    "incoming_webhook": {
        "url": "https://hooks.slack.com/TXXXXX/BXXXXX/XXXXXXXXXX",
        "channel": "#channel-it-will-post-to",
        "configuration_url": "https://teamname.slack.com/services/BXXXXX"
    },
    "bot":{
        "bot_user_id":"UTTTTTTTTTTR",
        "bot_access_token":"xoxb-XXXXXXXXXXXX-TTTTTTTTTTTTTT"
    }
}
```
*/
func (mess *SlackMessenger) OauthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = gotils.With(ctx, "messenger", messengers.ServiceSlack)
	ctx = gotils.With(ctx, "function", "OauthHandler")
	code := r.URL.Query().Get("code")
	oresp, err := slack.GetOAuthV2ResponseContext(ctx, http.DefaultClient, mess.clientID, mess.clientSecret, code,
		mess.oauthRedirectURI)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(fmt.Sprintf("Invalid authorization code from slack: %v", err)))
		return
	}
	//check if team already exists and update any tokens if they've changed
	ctx = gotils.With(ctx, "teamID", oresp.Team.ID)
	team, err := datastore.GetTeam(ctx, mess.db, messengers.ServiceSlack, oresp.Team.ID)
	if err != nil {
		if err != gotils.ErrNotFound {
			gotils.LogBeta(ctx, "error", "error getting team", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// if not found, we make the team
		team = &models.Team{
			Service:        messengers.ServiceSlack,
			ServiceID:      oresp.Team.ID,
			Name:           oresp.Team.Name,
			AccessToken:    oresp.AccessToken,
			BotUserID:      oresp.BotUserID,
			BotAccessToken: oresp.AccessToken,
		}
		team, err = datastore.SaveTeam(ctx, mess.db, team)
		if err != nil {
			gotils.LogBeta(ctx, "error", "couldn't save team! %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		// update tokens in case they have changed
		changed := false
		if team.Name != oresp.Team.Name {
			team.Name = oresp.Team.Name
		}
		if team.AccessToken != oresp.AccessToken {
			team.AccessToken = oresp.AccessToken
			changed = true
		}
		if team.BotUserID != oresp.BotUserID {
			team.BotUserID = oresp.BotUserID
			changed = true
		}
		if team.BotAccessToken != oresp.AccessToken {
			team.BotAccessToken = oresp.AccessToken
			changed = true
		}
		if changed {
			team, err = datastore.SaveTeam(ctx, mess.db, team)
			if err != nil {
				gotils.LogBeta(ctx, "error", "couldn't update team! %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
	// alright, think we're good
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("You're all good! Bot was added to %v. Return to Slack to interact with the bot.", team.Name)))
}
