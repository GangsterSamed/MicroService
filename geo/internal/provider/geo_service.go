package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ekomobile/dadata/v2/api/suggest"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/internal/models"
)

func (g *GeoService) AddressSearch(args *models.SearchArgs, reply *[]*models.Address) error {
	rawRes, err := g.api.Address(context.Background(), &suggest.RequestParams{Query: args.Query})
	if err != nil {
		return err
	}

	for _, r := range rawRes {
		*reply = append(*reply, &models.Address{
			City:   r.Data.City,
			Street: r.Data.Street,
			House:  r.Data.House,
			Lat:    r.Data.GeoLat,
			Lon:    r.Data.GeoLon,
		})
	}

	if len(*reply) == 0 {
		return fmt.Errorf("no addresses found")
	}

	return nil
}

func (g *GeoService) GeoCode(args *models.GeoArgs, reply *[]*models.Address) error {
	httpClient := &http.Client{}
	var data = strings.NewReader(fmt.Sprintf(`{"lat": %s, "lon": %s}`, args.Lat, args.Lng))

	req, err := http.NewRequest(
		"POST",
		"https://suggestions.dadata.ru/suggestions/api/4_1/rs/geolocate/address",
		data,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", g.apiKey))

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var geoCode struct {
		Suggestions []struct {
			Data struct {
				City   string `json:"city"`
				Street string `json:"street"`
				House  string `json:"house"`
				GeoLat string `json:"geo_lat"`
				GeoLon string `json:"geo_lon"`
			} `json:"data"`
		} `json:"suggestions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geoCode); err != nil {
		return err
	}

	for _, r := range geoCode.Suggestions {
		if r.Data.City == "" || r.Data.Street == "" {
			continue
		}

		*reply = append(*reply, &models.Address{
			City:   r.Data.City,
			Street: r.Data.Street,
			House:  r.Data.House,
			Lat:    r.Data.GeoLat,
			Lon:    r.Data.GeoLon,
		})
	}

	if len(*reply) == 0 {
		return fmt.Errorf("no geodata found")
	}

	return nil
}
