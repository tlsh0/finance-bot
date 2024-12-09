package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Token string
}

type Service struct {
	config Config
}

func New() (*Service, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	token := os.Getenv("token")

	s := &Service{
		Config{
			Token: token,
		},
	}

	return s, nil
}

func (s *Service) Token() string {
	return s.config.Token
}
