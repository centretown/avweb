package runtime

import (
	"encoding/json"
	"html/template"
	"os"
	"testing"
)

func TestWeather(t *testing.T) {
	var (
		daily   WeatherDaily
		hourly  WeatherHourly
		current WeatherCurrent
	)

	config := testConfig(t)

	testFile(t, "testdata/daily.json", &daily)
	testFile(t, "testdata/hourly.json", &hourly)
	testFile(t, "testdata/current.json", &current)

	tmpl, err := template.ParseGlob("../www/*.html")
	if err != nil {
		t.Fatal(err)
	}

	tmp := tmpl.Lookup("weather.current")

	for _, l := range config.Locations {
		l.WeatherCurrent = &current
	}

	rt := &Runtime{
		Location:  config.Locations[0],
		Locations: config.Locations,
	}
	data := &WeatherFormData{
		Codes:   WeatherCodes,
		Data:    rt.Locations,
		Runtime: rt}

	data.Action = config.Actions["weather_current"]
	rt.Location.WeatherCurrent = &current
	err = tmp.Execute(os.Stdout, data)
	if err != nil {
		t.Fatal(err)
	}
}

func testConfig(t *testing.T) *Config {
	filename := "testdata/config.json"
	config := NewConfig()
	err := config.Read(filename)
	if err != nil {
		t.Fatal("Read Configuration", filename, err)
	}
	return config
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
