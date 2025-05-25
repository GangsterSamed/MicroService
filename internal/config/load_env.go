package config

import (
	"errors"
	"github.com/caarlos0/env/v9"
)

// LoadAuthConfig загружает конфигурацию для сервиса auth
func LoadAuthConfig() (*AuthConfig, error) {
	var cfg AuthConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, errors.New("JWT_SECRET must be at least 32 characters long")
	}
	return &cfg, nil
}

// LoadGeoConfig загружает конфигурацию для сервиса geo
func LoadGeoConfig() (*GeoConfig, error) {
	var cfg GeoConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	if cfg.DadataAPIKey == "" || cfg.DadataSecretKey == "" {
		return nil, errors.New("DADATA_API_KEY and DADATA_SECRET_KEY are required")
	}
	return &cfg, nil
}

// LoadUserConfig загружает конфигурацию для сервиса user
func LoadUserConfig() (*UserConfig, error) {
	var cfg UserConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoadProxyConfig загружает конфигурацию для сервиса proxy
func LoadProxyConfig() (*ProxyConfig, error) {
	var cfg ProxyConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
