package middleware

import (
	"fmt"

	"github.com/seargo/seargo/internal/config"
)

// ValidateSecretKey checks that the server's secret_key is not the default value
// when running in production mode (debug = false).
func ValidateSecretKey(cfg *config.Config) error {
	if cfg.General.Debug {
		return nil
	}
	if cfg.Server.SecretKey == "" || cfg.Server.SecretKey == "ultrasecretkey" {
		return fmt.Errorf("server.secret_key must be changed from the default %q in production mode", cfg.Server.SecretKey)
	}
	return nil
}
