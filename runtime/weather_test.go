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
	testFile(t, "testdata/hourly2.json", &hourly)
	testFile(t, "testdata/current.json", &current)

	for i, l := range config.Locations {
		l.WeatherDaily = &daily
		l.WeatherHourly = &hourly
		t.Log("BuildHourlyProperties", i)
		l.BuildHourlyProperties(i)
		buf, _ := json.MarshalIndent(l.HourlyProperties, "", "  ")
		t.Log(string(buf))
		l.WeatherCurrent = &current
	}

	testCurrent(t, config, &hourly)
	return
}
func testCurrent(t *testing.T, config *Config, current *WeatherHourly) {
	tmpl, err := template.ParseGlob("../www/*.html")
	if err != nil {
		t.Fatal(err)
	}

	tmp := tmpl.Lookup("weather.hourly.properties")

	rt := &Runtime{
		Location:  config.Locations[0],
		Locations: config.Locations,
	}
	// data := &WeatherFormData{
	// 	Codes:   WeatherCodes,
	// 	Data:    rt.Locations,
	// 	Runtime: rt}

	rt.Location.WeatherHourly = current
	err = tmp.Execute(os.Stdout, rt.Location)
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
