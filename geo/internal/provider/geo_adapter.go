package provider

import "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/internal/models"

// Адаптер для интерфейса GeoProvider
type geoProviderAdapter struct {
	service *GeoService
}

func NewGeoProviderAdapter(service *GeoService) GeoProvider {
	return &geoProviderAdapter{service: service}
}

func (a *geoProviderAdapter) AddressSearch(input string) ([]*models.Address, error) {
	args := &models.SearchArgs{Query: input}
	var reply []*models.Address
	err := a.service.AddressSearch(args, &reply)
	return reply, err
}

func (a *geoProviderAdapter) GeoCode(lat, lng string) ([]*models.Address, error) {
	args := &models.GeoArgs{Lat: lat, Lng: lng}
	var reply []*models.Address
	err := a.service.GeoCode(args, &reply)
	return reply, err
}
