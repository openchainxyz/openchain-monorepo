package signature_database_srv

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/openchainxyz/openchainxyz-monorepo/internal/solidity"
	"github.com/openchainxyz/openchainxyz-monorepo/services/signature-database-srv/client"
	log "github.com/sirupsen/logrus"
	"strings"
)

func (s *Service) importRaw(data client.ImportRequest) (client.ImportResponse, error) {
	response := client.NewImportResponse()

	var err error
	for _, typ := range client.SignatureTypes() {
		response[typ], err = s.importRawType(typ, data[typ])
		if err != nil {
			return nil, fmt.Errorf("failed to save signatures to db: %w", err)
		}
	}

	return response, nil
}

func (s *Service) importRawType(typ client.SignatureType, input []string) (*client.ImportResponseDetails, error) {
	var pending []string
	var invalid []string
	for _, text := range input {
		if solidity.VerifySignature(text) {
			pending = append(pending, text)
		} else {
			invalid = append(invalid, text)
		}
	}

	resp, err := s.db.SaveSignatures(typ, pending)
	if err != nil {
		return nil, err
	}

	if s.discord != nil {
		if err := s.notifyDiscord(typ, resp); err != nil {
			log.WithError(err).Errorf("failed to notify discord of import")
		}
	}

	resp.Invalid = invalid
	return resp, nil
}

func (s *Service) notifyDiscord(typ client.SignatureType, resp *client.ImportResponseDetails) error {
	var imported []string
	for _, hash := range resp.Imported {
		imported = append(imported, hash)
	}

	sigs, err := s.db.LoadSignatures(typ, imported)
	if err != nil {
		return err
	}

	var parts []string
	for sig, data := range sigs {
		if len(data) <= 1 {
			continue
		}

		var names []string
		for _, name := range data {
			names = append(names, fmt.Sprintf("`%s`", name.Name))
		}

		parts = append(parts, fmt.Sprintf("`%s`: %s", sig, strings.Join(names, ", ")))
	}

	if len(parts) > 0 {
		if _, err := s.discord.Session.ChannelMessageSendComplex(s.config.DiscordChannel, &discordgo.MessageSend{
			Content: fmt.Sprintf("Imported the following duplicate signatures:\n%s", strings.Join(parts, "\n")),
		}); err != nil {
			return err
		}
	}

	return nil
}
