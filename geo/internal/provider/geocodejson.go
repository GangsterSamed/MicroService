package provider

import "encoding/json"

type GeoCode struct {
	Suggestions []Suggestion `json:"suggestions"`
}

type Suggestion struct {
	Value             string `json:"value"`
	UnrestrictedValue string `json:"unrestricted_value"`
	Data              Data   `json:"data"`
}

type Data struct {
	City    string `json:"city"`
	Street  string `json:"street"`
	House   string `json:"house"`
	GeoLat  string `json:"geo_lat"`
	GeoLon  string `json:"geo_lon"`
	FiasID  string `json:"fias_id"`
	KladrID string `json:"kladr_id"`
	QcGeo   string `json:"qc_geo"`
}

func UnmarshalGeoCode(data []byte) (GeoCode, error) {
	var r GeoCode
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *GeoCode) Marshal() ([]byte, error) {
	return json.Marshal(r)
}
