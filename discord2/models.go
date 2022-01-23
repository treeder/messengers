package discord

type ChannelRequest struct {
	RecipientID string `json:"recipient_id"`
}

type Channel struct {
	ID   string `json:"id"`
	Type int    `json:"type"`
}
