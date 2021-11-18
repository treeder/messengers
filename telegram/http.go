package telegram

import (
	"bytes"
	"encoding/json"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/treeder/gotils/v2"
)

func (mess *TelegramMessenger) HandleEventHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = gotils.With(ctx, "messenger", "telegram")
	// parse update object and pass it along
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	// body := buf.String()
	// gotils.L(ctx).Sugar().Debug("BODY: ", string(body))
	update := &tgbotapi.Update{}
	err := json.Unmarshal(buf.Bytes(), update)
	if err != nil {
		gotils.LogBeta(ctx, "error", "error parsing Telegram Update... wtf? %v", (err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// fmt.Printf("handler slice 2: %+v\n", mess.MessageHandlers)
	if update.Message == nil {
		w.Write([]byte(http.StatusText(http.StatusOK)))
		return
	}
	if mess.Client().Self.ID == update.Message.From.ID {
		// skip bots own messages
		w.Write([]byte(http.StatusText(http.StatusOK)))
		return
	}
	if update.Message.From.IsBot {
		// skip bot messages
		w.Write([]byte(http.StatusText(http.StatusOK)))
		return
	}
	msg := NewInMsg(ctx, update)
	for _, h := range mess.MessageHandlers {
		// fmt.Println("handling message", h)
		h.HandleMessage(ctx, mess, msg)
	}
	w.WriteHeader(http.StatusOK)
}
