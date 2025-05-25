package models

type Address struct {
	City   string `json:"city"`
	Street string `json:"street"`
	House  string `json:"house"`
	Lat    string `json:"lat"`
	Lon    string `json:"lon"`
}

// RequestAddressSearch описывает запрос на поиск адреса
type RequestAddressSearch struct {
	Query string `json:"query"`
}

// RequestAddressGeocode описывает запрос на геокодирование
type RequestAddressGeocode struct {
	Lat string `json:"lat"`
	Lng string `json:"lng"`
}

type ResponseAddress struct {
	Addresses []*Address `json:"addresses"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SearchArgs struct {
	Query string `json:"query"`
}

type GeoArgs struct {
	Lat string `json:"lat"`
	Lng string `json:"lng"`
}
