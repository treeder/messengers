package models

import "github.com/treeder/firetils"

type Team struct {
	firetils.Firestored
	firetils.Timestamped
	Service        string `firestore:"service"`    // only a team service
	ServiceID      string `firestore:"service_id"` // The teams ID for the service
	Name           string `firestore:"name"`
	AccessToken    string `firestore:"access_token"`
	BotUserID      string `firestore:"bot_user_id"`
	BotAccessToken string `firestore:"bot_access_token"`
}
