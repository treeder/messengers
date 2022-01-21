package discord

import (
	"context"
	"encoding/hex"

	"github.com/treeder/messengers"
)

// New creates new discord bot client
// todo: do we need the clientid and secret??  remove if not
func New(ctx context.Context, clientID, publicKey, clientSecret, token string) (*DiscordMessenger, error) {
	hexEncodedDiscordPubkey := publicKey
	discordPubkey, err := hex.DecodeString(hexEncodedDiscordPubkey)
	if err != nil {
		return nil, err
	}

	mess := &DiscordMessenger{
		BaseMessenger:    &messengers.BaseMessenger{},
		ClientID:         clientID,
		PublicKey:        publicKey,
		decodedPublicKey: discordPubkey,
		baseCtx:          ctx,
	}
	return mess, nil
}
