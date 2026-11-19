package cli

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Issuer   string `env:"AUTH_ISSUER,required,notEmpty"`
	Audience string `env:"AUTH_AUDIENCE,required,notEmpty"`
	ClientID string `env:"CLI_AUTH0_CLIENT_ID,required,notEmpty"`
	APIURL   string `env:"CLI_API_URL" envDefault:"http://localhost:8080"`
}

func LoadConfig() (Config, error) {
	// Dev convenience: the repo's .env. Real env vars win.
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env file: %w", err)
	}
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("read environment: %w", err)
	}
	return cfg, nil
}
