package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/treeder/gotils/v2"
	"github.com/treeder/messengers"
)

var (
	// todo: REMOVE THESE GLOBALS!
	discordClient *discordgo.Session
	discordApp    *discordgo.Application
	clientID      string
	messGlobal    *DiscordMessenger
)

// New creates new discord bot client
// todo: do we need the clientid and secret??  remove if not
func New(ctx context.Context, clientIDIn, clientSecret, token string) (*DiscordMessenger, error) {
	clientID = clientIDIn

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return nil, err
	}
	// discordApp, err = dg.Application(clientID)
	// if err != nil {
	// 	fmt.Println("error getting Discord application,", err)
	// 	return nil, err
	// }
	// fmt.Printf("APP: %+v", discordApp)
	mess := &DiscordMessenger{
		BaseMessenger: &messengers.BaseMessenger{},
		sess:          dg,
		ClientID:      clientID,
		baseCtx:       ctx,
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(mess.messageCreate)

	// Open a websocket connection to Discord and begin listening.
	fmt.Println("Starting discord bot")
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return nil, err
	}
	discordClient = dg
	messGlobal = mess
	return mess, nil
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func (mess *DiscordMessenger) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// ex: msg: &{ID:488926884802068484 ChannelID:306937786399719434 Content:this is a **test** yo Timestamp:2018-09-11T04:20:53.159000+00:00 EditedTimestamp: MentionRoles:[] Tts:false MentionEveryone:false Author:treeder#6967 Attachments:[] Embeds:[] Mentions:[] Reactions:[] Type:0}
	ctx := mess.baseCtx // since discordgo doesn't support ctx yet, adding it here - https://github.com/bwmarrin/discordgo/pull/369
	rid, _ := gonanoid.New()
	ctx = gotils.With(ctx, "request_id", rid)
	defer func() {
		if r := recover(); r != nil {
			gotils.LogBeta(ctx, "error", "Recovered in discord.messageCreate(): %v", r)
			// sendError(ctx, update.Message, gotils.C(ctx).Errorf("Server error occurred. We'll fix this asap! Please try again later."))
		}
	}()

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	t := m.Content
	if !strings.HasPrefix(t, "/") {
		// ignore if not a slash command
		return
	}
	cmd, tsplit := messengers.ParseCommand(ctx, t)

	// todo: do the whole same thing as telegram
	u, err := s.User(m.Author.ID)
	if err != nil {
		fmt.Println("error getting user,", err)
		return
	}
	fmt.Printf("user: %+v\n", u)
	fmt.Println(u.ID, u.Email, u.Username, u.Discriminator)

	// TODO: Save user as u.ID and username use: u.String() which is username+discriminator

	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}

	msg := &InMsg{
		Msg:   m.Message,
		cmd:   cmd,
		split: tsplit,
		sess:  s,
		ctx:   ctx,
	}
	// sess should be moved into message maybe and use global Messenger? little confusing
	// mess := &DiscordMessenger{
	// 	sess:     s,
	// 	ClientID: clientID,
	// }

	for _, h := range mess.MessageHandlers {
		h.HandleMessage(ctx, mess, msg)
	}
}
