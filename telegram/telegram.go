package telegram

import (
	"context"
	"encoding/base64"

	"cloud.google.com/go/firestore"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/treeder/gotils/v2"
	"github.com/treeder/messengers"
)

func New(ctx context.Context, db *firestore.Client, apiKey, rsaKeyForPassport string) (*TelegramMessenger, error) {
	// infra = infraIn
	// apiKey = tgApiKey
	var err error
	if apiKey == "" {
		// gotils.L(ctx).Fatal("Please set Telegram API KEY env var: TG_API_KEY")
		return nil, gotils.C(ctx).Errorf("Telegram API KEY required")
	}
	data, err := base64.StdEncoding.DecodeString(rsaKeyForPassport)
	if err != nil {
		return nil, gotils.C(ctx).Errorf("FATAL Invalid TG_RSA_KEY: %v", err)
	}
	// fmt.Println("RSA_KEY:", string(data))

	bot, err := tgbotapi.NewBotAPI(apiKey)
	if err != nil {
		return nil, gotils.C(ctx).Errorf("FATAL on NewBotAPI: %v", err)
	}

	// Turn this on to see all telegram raw messages
	// bot.Debug = true

	mess := &TelegramMessenger{
		BaseMessenger:     &messengers.BaseMessenger{},
		tg:                bot,
		rsaKeyForPassport: data,
		apiKey:            apiKey,
	}

	// u := tgbotapi.NewUpdate(0)
	// u.Timeout = 60
	// updates, err := bot.GetUpdatesChan(u)
	// go func() {
	// 	for update := range updates {
	// 		// log.Printf("UPDATE: %+v", update)
	// 		go mess.handleUpdate(ctx, update)
	// 	}
	// }()

	// messGlobal = mess
	return mess, nil
}
