package models

import (
	"strconv"

	"github.com/treeder/firetils"
)

// ServiceUser user info for each service, eg: Telegram
// can't be a sub collection because you wouldn't know the main User ID (above)
type ServiceUser struct {
	firetils.Firestored
	firetils.IDed
	firetils.Timestamped

	// ID the user ID for the service, NOT the firestore ID
	ID string `firestore:"id"`
	// Service name, eg: "telegram" Must used ID and Service together for uniqueness
	Service string `firestore:"service"`
	// Username  on the service, eg: @treeder on tg
	Username string `firestore:"username"`
	// UsernameLower lower cased version for querying
	UsernameLower string `firestore:"username_lower"`
	// DMChatID is for services like hangouts that have a room ID for DMs
	DMChatID string `firestore:"chat_id"`

	// UserID maps to the `users` collection
	UserID string `firestore:"user_id"`

	// User configured settings/preferences
	Settings map[string]interface{} `firestore:"settings"`

	// CountTx total count of sent transactions
	CountTx int64 `firestore:"count_tx"`
	// SumSent total ever sent in wei
	SumSent string `firestore:"sum_sent"`
	
	OauthToken  string `firestore:"oauthToken" json:"-"`
	OauthSecret string `firestore:"oauthSecret" json:"-"`
}

func (su *ServiceUser) Network() string {
	if su.Settings == nil {
		return "mainnet"
	}
	if val, ok := su.Settings["network"]; ok {
		return val.(string)
	}
	return "mainnet"
}

func (su *ServiceUser) IntID() int64 {
	i, err := strconv.ParseInt(su.ID, 10, 64)
	if err != nil {
		panic(err)
	}
	return i
}

func (su *ServiceUser) ServicePlusID() string {
	return su.Service + "-" + su.ID
}
