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

	testFile(t, "testdata/daily.json", &daily)
	testFile(t, "testdata/hourly.json", &hourly)
	testFile(t, "testdata/current.json", &current)

	tmpl, err := template.ParseGlob("../www/current.html")
	if err != nil {
		t.Fatal(err)
	}

	tmp := tmpl.Lookup("weather.current.titles")
	location := &Location{
		City:           "TestCity",
		WeatherCurrent: &current,
	}
	err = tmp.Execute(os.Stdout, location)
	if err != nil {
		t.Fatal(err)
	}
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
