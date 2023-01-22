package service

type Config struct {
	HttpPort int `def:"34887" env:"PORT"`
}

type Service struct {
	config *Config
}

func New(config *Config) (*Service, error) {
	return &Service{config: config}, nil
}

func (s *Service) Start() error {
	go s.startServer()

	return nil
}
