package provider

import (
	"github.com/ekomobile/dadata/v2/api/suggest"
	"github.com/ekomobile/dadata/v2/client"
	"net/url"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/internal/models"
)

type GeoService struct {
	api       *suggest.Api
	apiKey    string
	secretKey string
}

type GeoProvider interface {
	AddressSearch(input string) ([]*models.Address, error)
	GeoCode(lat, lng string) ([]*models.Address, error)
}

func NewGeoService(apiKey, secretKey string) *GeoService {
	var err error
	endpointUrl, err := url.Parse("https://suggestions.dadata.ru/suggestions/api/4_1/rs/")
	if err != nil {
		return nil
	}

	creds := client.Credentials{
		ApiKeyValue:    apiKey,
		SecretKeyValue: secretKey,
	}

	api := suggest.Api{
		Client: client.NewClient(endpointUrl, client.WithCredentialProvider(&creds)),
	}

	return &GeoService{
		api:       &api,
		apiKey:    apiKey,
		secretKey: secretKey,
	}
}
