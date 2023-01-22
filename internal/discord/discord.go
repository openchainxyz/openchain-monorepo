package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	Session *discordgo.Session
}

func New(token string) (*Client, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("failed to create discord client: %w", err)
	}

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.WithFields(log.Fields{
			"username": fmt.Sprintf("%s#%s", s.State.User.Username, s.State.User.Discriminator),
		}).Infof("logged in to discord")

		// register application commands if any
	})

	if err := session.Open(); err != nil {
		return nil, fmt.Errorf("failed to open session: %w", err)
	}

	return &Client{
		Session: session,
	}, nil
}
