package config

import (
	"fmt"

	"github.com/caarlos0/env/v9"
)

// LoadAuthConfig загружает конфигурацию для сервиса auth
func LoadAuthConfig() (*AuthConfig, error) {
	var cfg AuthConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse env config: %w", err)
	}

	// Валидируем конфигурацию
	if err := validateAuthConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// LoadGeoConfig загружает конфигурацию для сервиса geo
func LoadGeoConfig() (*GeoConfig, error) {
	var cfg GeoConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse env config: %w", err)
	}

	// Валидируем конфигурацию
	if err := validateGeoConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// LoadUserConfig загружает конфигурацию для сервиса user
func LoadUserConfig() (*UserConfig, error) {
	var cfg UserConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse env config: %w", err)
	}

	// Валидируем конфигурацию
	if err := validateUserConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// LoadProxyConfig загружает конфигурацию для сервиса proxy
func LoadProxyConfig() (*ProxyConfig, error) {
	var cfg ProxyConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse env config: %w", err)
	}

	// Валидируем конфигурацию
	if err := validateProxyConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}
