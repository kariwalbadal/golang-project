package geocode

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
)

const AccessToken string = "pk.9d6311f1f89aaacc09db7c9a0866bc2c"

type RevGeocodeResponse struct {
	PlaceID     string `json:"place_id"`
	Licence     string `json:"licence"`
	OsmType     string `json:"osm_type"`
	OsmID       string `json:"osm_id"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
	Address     struct {
		School        string `json:"school"`
		Road          string `json:"road"`
		Neighbourhood string `json:"neighbourhood"`
		Suburb        string `json:"suburb"`
		CityDistrict  string `json:"city_district"`
		City          string `json:"city"`
		County        string `json:"county"`
		StateDistrict string `json:"state_district"`
		State         string `json:"state"`
		Postcode      string `json:"postcode"`
		Country       string `json:"country"`
		CountryCode   string `json:"country_code"`
	} `json:"address"`
	Boundingbox []string `json:"boundingbox"`
}

func FloatToString(arg float64) string {
	return strconv.FormatFloat(arg, 'f', 6, 64)
}

func RevGeocodeState(latitude float64, longitude float64) string {

	res, requestError := http.Get("https://eu1.locationiq.com/v1/reverse.php?key=" + AccessToken + "&lat=" + FloatToString(latitude) + "&lon=" + FloatToString(longitude) + "&format=json")

	response := RevGeocodeResponse{}

	if requestError != nil {
		return requestError.Error()
	}

	defer res.Body.Close()

	bodyBytes, readBodyError := ioutil.ReadAll(res.Body)
	if readBodyError != nil {
		return readBodyError.Error()
	}

	parseError := json.Unmarshal(bodyBytes, &response)

	if parseError != nil {
		return parseError.Error()
	}

	return response.Address.State

}
