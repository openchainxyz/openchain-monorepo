package signature_database_srv

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/openchainxyz/openchainxyz-monorepo/internal/discord"
	"github.com/openchainxyz/openchainxyz-monorepo/services/signature-database-srv/database"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type Config struct {
	DatabaseHost     string `def:"127.0.0.1" env:"DB_HOST"`
	DatabasePort     int    `def:"5432" env:"DB_PORT"`
	DatabaseName     string `def:"postgres" env:"DB_NAME"`
	DatabaseUser     string `def:"ethereum" env:"DB_USER"`
	DatabasePassword string `def:"ethereum" env:"DB_PASS"`
	HttpPort         int    `def:"34887" env:"PORT"`
	DiscordBotToken  string `env:"DISCORD_BOT_TOKEN"`
	DiscordChannel   string `env:"DISCORD_CHANNEL"`

	DataDumpDir string `env:"DATA_DUMP_DIR"`
}

type Service struct {
	config  *Config
	db      *database.Database
	discord *discord.Client

	canonicalSignaturesLock        sync.RWMutex
	canonicalSignatures            map[string]string
	lastCanonicalSignaturesRefresh time.Time

	dataExportLock     sync.Mutex
	dataExportPath     string
	lastDataExportTime time.Time
}

func New(config *Config) (*Service, error) {
	db, err := database.New(config.DatabaseHost, config.DatabasePort, config.DatabaseName, config.DatabaseUser, config.DatabasePassword)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	service := &Service{
		config: config,
		db:     db,

		canonicalSignaturesLock: sync.RWMutex{},
		canonicalSignatures:     make(map[string]string),

		dataExportLock: sync.Mutex{},
	}

	if config.DiscordBotToken != "" {
		discordClient, err := discord.New(config.DiscordBotToken)
		if err != nil {
			return nil, fmt.Errorf("failed to create discord bot: %w", err)
		}
		service.discord = discordClient
	}

	if err := service.loadCanonicalSignatures(); err != nil {
		return nil, fmt.Errorf("failed to load canonical signatures: %w", err)
	}

	return service, nil
}

func (s *Service) Start() error {
	go s.startServer()
	go s.runTasks()

	return nil
}

func (s *Service) loadCanonicalSignatures() error {
	s.canonicalSignaturesLock.Lock()
	lastRefresh := s.lastCanonicalSignaturesRefresh
	s.canonicalSignaturesLock.Unlock()

	if time.Since(lastRefresh) < 1*time.Hour {
		return fmt.Errorf("refreshing too soon")
	}

	resp, err := http.Get("https://raw.githubusercontent.com/openchainxyz/canonical-signatures/main/canonical.yaml")
	if err != nil {
		return fmt.Errorf("failed to fetch canonical signatures: %w", err)
	}

	var output map[string]struct {
		Signature string `yaml:"signature"`
		Source    string `yaml:"source"`
	}

	if err := yaml.NewDecoder(resp.Body).Decode(&output); err != nil {
		return fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	newCanonicalSignatures := make(map[string]string)
	for k, v := range output {
		newCanonicalSignatures[k] = v.Signature
	}

	s.canonicalSignaturesLock.Lock()
	s.canonicalSignatures = newCanonicalSignatures
	s.lastCanonicalSignaturesRefresh = time.Now()
	s.canonicalSignaturesLock.Unlock()

	return nil
}

func (s *Service) exportData() error {
	s.dataExportLock.Lock()
	lastExportTime := s.lastDataExportTime
	s.dataExportLock.Unlock()

	if time.Since(lastExportTime) < 24*time.Hour {
		return fmt.Errorf("exporting too soon")
	}

	newPath := path.Join(s.config.DataDumpDir, uuid.New().String()+".txt")

	f, err := os.Create(newPath)
	if err != nil {
		return err
	}

	if err := s.db.ExportData(f); err != nil {
		f.Close()
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	s.dataExportLock.Lock()
	lastPath := s.dataExportPath
	s.dataExportPath = newPath
	s.lastDataExportTime = time.Now()
	s.dataExportLock.Unlock()

	os.Remove(lastPath)

	return nil
}
func (s *Service) runTasks() {
	ticker := time.NewTicker(24 * time.Hour)
	for ; true; <-ticker.C {
		if err := s.loadCanonicalSignatures(); err != nil {
			log.WithError(err).Errorf("failed to load canonical signatures")
		} else {
			log.Info("successfully refreshed canonical signatures")
		}
		if err := s.exportData(); err != nil {
			log.WithError(err).Errorf("failed to export data")
		} else {
			log.Info("successfully exported data")
		}
	}
}
