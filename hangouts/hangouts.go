package hangouts

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/treeder/gotils/v2"
	"github.com/treeder/messengers"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	chat "google.golang.org/api/chat/v1"
)

const (
	// subName = "hangouts-sub"

	// don't change
	ChatScope = "https://www.googleapis.com/auth/chat.bot"
)

func New(ctx context.Context, db *firestore.Client, botID string, serviceAccountEncoded string) (*HangoutsMessenger, error) {
	ctx = gotils.With(ctx, "messenger", "hangouts")

	// STOPPED USING PUBSUB, WAY TOO COMPLICATED. JUST USING WEBHOOKS NOW
	// pubsubClient, err := pubsub.NewClient(ctx, infra.GProjectID)
	// if err != nil {
	// 	gotils.LogBeta(ctx, "error", "Failed to get create pubsubClient", err)
	// 	return nil, err
	// }

	// todo: can use below once published as a global bot I believe
	// gclient, err := google.DefaultClient(ctx, "https://www.googleapis.com/auth/chat.bot")
	// if err != nil {
	// 	gotils.LogBeta(ctx, "error", "Failed to create google.DefaultClient", err)
	// 	return err
	// }
	// creds, err := google.CredentialsFromJSON(ctx, data, chatScope)
	// if err != nil {
	// 	gotils.LogBeta(ctx, "error", "Failed to create google.DefaultClient", err)
	// 	return nil, err
	// }
	if serviceAccountEncoded == "" {
		// gotils.L(ctx).Warn("HANGOUTS_ACCOUNT not set, won't be starting hangouts support.")
		return nil, gotils.C(ctx).Errorf("HANGOUTS_KEY required")
	}
	serviceAccountJSON, err := base64.StdEncoding.DecodeString(serviceAccountEncoded)
	if err != nil {
		gotils.LogBeta(ctx, "warn", "Failed to read Hangouts chat credentials from env, won't be starting hangouts support: %v", (err))
		return nil, nil
	}
	// fmt.Println("HANGOUTS_ACCOUNT:", string(serviceAccountJSON))
	creds, err := google.CredentialsFromJSON(ctx, serviceAccountJSON, ChatScope)
	if err != nil {
		gotils.LogBeta(ctx, "warn", "Failed to parse Hangouts chat credentials, won't be starting hangouts support: %v", (err))
		return nil, nil
	}

	gclient := oauth2.NewClient(ctx, creds.TokenSource)
	hchat, err := chat.New(gclient)
	if err != nil {
		gotils.LogBeta(ctx, "warn", "Failed to get create hchat service: %v", (err))
		return nil, err
	}
	mess := &HangoutsMessenger{
		BaseMessenger: &messengers.BaseMessenger{},
		hchat:         hchat,
		_botID:        botID,
	}
	gotils.LogBeta(ctx, "info", "Starting hangouts bot")
	// go startReceiving(ctx, pubsubClient, hchat)
	return mess, nil
}

// func startReceiving(ctx context.Context, pubsubClient *pubsub.Client, hchat *chat.Service) error {
// 	// Consume 10 messages.
// 	// var mu sync.Mutex
// 	// received := 0
// 	sub := pubsubClient.Subscription(subName)
// 	// if need to stop getting messages, use cancel below
// 	cctx, _ := context.WithCancel(ctx)
// 	err := sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
// 		msg.Ack()
// 		fmt.Printf("[HANGOUT] Got message: %v\n", string(msg.Data))
// 		// mu.Lock()
// 		// defer mu.Unlock()
// 		// received++
// 		// if received == 10 {
// 		// 	cancel()
// 		// }
// 		// hchat.
// 		err := handleEvent(ctx, msg)
// 		if err != nil {
// 			return
// 		}
// 	})
// 	if err != nil {
// 		gotils.LogBeta(ctx, "error", "error getting message from pubsub", err)
// 		return err
// 	}
// 	return nil
// }
func (mess *HangoutsMessenger) AddHandler(ctx context.Context, h messengers.MessageHandler) {
	mess.MessageHandlers = append(mess.MessageHandlers, h)
}

func (mess *HangoutsMessenger) HandleEventHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body := r.Body
	event := &chat.DeprecatedEvent{}
	decoder := json.NewDecoder(body)
	err := decoder.Decode(event)
	if err != nil {
		gotils.LogBeta(ctx, "error", "error unmarshaling message from HTTP: %v", (err))
		http.Error(w, "error unmarshaling message from HTTP", http.StatusBadRequest)
		return
	}
	rmsg, err := handleEvent2(ctx, mess, event)
	if err != nil {
		switch v := err.(type) {
		case gotils.HTTPError:
			http.Error(w, err.Error(), v.Code())
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rmsg != nil {
		b, err := json.Marshal(rmsg)
		if err != nil {
			gotils.LogBeta(ctx, "error", "error marshalling ResponseMsg: %v", (err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(b)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func handleEvent2(ctx context.Context, mess *HangoutsMessenger, event *chat.DeprecatedEvent) (*messengers.ResponseMsg, error) {
	switch event.Type {
	case "ADDED_TO_SPACE":
		if event.Type == "DM" {
			event.Message = &chat.Message{Text: "start"}
			msg := NewMsg(ctx, mess, event)
			for _, h := range mess.MessageHandlers {
				h.HandleMessage(ctx, mess, msg)
			}
			// rmsg, err := core.ProcessMessage2(ctx, mess, msg)
			// if err != nil {
			// 	if err == utils.ErrUnknownCommand {
			// 		rmsg = &messengers.ResponseMsg{Text: "Sorry, I don't know that command."}
			// 		return rmsg, nil
			// 	}
			// 	gotils.LogBeta(ctx, "error", "error from ProcessMessage", err)
			// 	mess.SendError(ctx, msg, err)
			// 	return nil, err
			// }
		}
	case "MESSAGE":
		msg := NewMsg(ctx, mess, event)
		for _, h := range mess.MessageHandlers {
			h.HandleMessage(ctx, mess, msg)
		}
		// rmsg, err := core.ProcessMessage2(ctx, mess, msg)
		// if err != nil {
		// 	if err == utils.ErrUnknownCommand {
		// 		rmsg = &messengers.ResponseMsg{Text: "Sorry, I don't know that command."}
		// 		return rmsg, nil
		// 	}
		// 	gotils.LogBeta(ctx, "error", "error from ProcessMessage", err)
		// 	// mess.SendError(ctx, msg, err)
		// 	return nil, err
		// }
		// return rmsg, nil
		// fmt.Println(res.)
	case "REMOVED_FROM_SPACE":
	default:
		// Do nothing
		return nil, nil
	}
	return nil, nil
}

// func trimPrefix(id string) string {
// 	return strings.TrimPrefix(id, "users/")
// }

// func addPrefix(id string) string {
// 	return "users/" + id
// }
