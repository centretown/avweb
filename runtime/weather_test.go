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
