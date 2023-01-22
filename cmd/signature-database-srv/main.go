package main

import (
	"fmt"
	"github.com/openchainxyz/openchainxyz-monorepo/internal/config"
	service "github.com/openchainxyz/openchainxyz-monorepo/services/signature-database-srv"
	log "github.com/sirupsen/logrus"
)

func run() error {
	var cfg service.Config
	if err := config.LoadConfig(&cfg); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	a, err := service.New(&cfg)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	if err := a.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	return nil
}

func main() {
	log.SetLevel(log.DebugLevel)
	if err := run(); err != nil {
		log.WithError(err).Fatalf("failed to run service")
	}
	<-make(chan struct{})
}
