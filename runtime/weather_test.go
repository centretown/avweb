package runtime

import (
	"encoding/json"
	"os"
	"testing"
)

func TestWeather(t *testing.T) {
	var (
		daily  WeatherDaily
		hourly WeatherHourly
	)

	testFile(t, "testdata/daily.json", &daily)
	testFile(t, "testdata/hourly.json", &hourly)
	daily.Log()

}

func testFile(t *testing.T, filename string, weather any) {
	var (
		err  error
		file *os.File
	)
	t.Log("processing", filename)
	file, err = os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	err = LoadWeather(file, weather)
	if err != nil {
		t.Fatal(err)
	}

	buf, err := json.MarshalIndent(weather, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(buf))
}

func TestMinutely(t *testing.T) {

	var (
		// trailer   = "&minutely_15=temperature_2m,apparent_temperature,precipitation_probability,precipitation,weather_code,wind_speed_10m"
		Locations = []*Location{
			{City: "Ottawa", Latitude: 45.40608984676536, Longitude: -75.68631292544273, Zone: "America%2FNew_York"},
			{City: "Timmins", Latitude: 48.485340413458964, Longitude: -81.33687676821839, Zone: "America%2FNew_York"},
			{City: "Kenora", Latitude: 49.77255342394314, Longitude: -94.48309874840045, Zone: "America%2FNew_York"},
		}
		loc = Locations[0]
	)
	err := loc.QueryMinutely()
	if err != nil {
		t.Fatal(err)
	}

	buf, err := json.Marshal(&loc.WeatherMinutely)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(buf))
}
