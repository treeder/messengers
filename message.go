package messengers

import "context"

// Message is the interface for all messages throughout the system
type Message interface {
	// the ID of the message
	ID() string
	TeamID() string // team in slack, guild in discord
	ChatID() string //ChatID is the room/group/channel/space
	FromID() string
	ThreadID() string
	FromUsername() string
	// IsPrivate says whether this is a private DM with the bot
	IsPrivate() bool
	// Mention spits out the appropriate string to tag the author
	Mention() string

	// Raw returns the original message from the API client
	Raw() interface{}
}

// IncomingMessage extends Message with more incoming (request) information
type IncomingMessage interface {
	Message

	// Cmd the first value in the incoming message, minus the slash
	Cmd() string
	Command() string
	IsSlashCommand() bool

	FullText() string
	// Split the incoming command split
	Split() []string

	// If a message comes in as a reply, these will be populated. Not all messengers support this.
	ReplyToMsgID() string
	ReplyToMsg() Message

	// Attachment[]

}

type MessageHandler interface {
	HandleMessage(ctx context.Context, mess Messenger, msg IncomingMessage)
}

type MessageHandlerRaw interface {
	HandleMessageRaw(ctx context.Context, mess Messenger, msg interface{})
}
