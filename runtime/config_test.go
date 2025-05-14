package runtime

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"testing"
)

func TestConfigRead(t *testing.T) {
	cfg := NewConfig()
	err := cfg.Read("testdata/config.json")
	if err != nil {
		t.Fatal(err)
	}
	buf, err := json.MarshalIndent(cfg, "", "  ")
	t.Log(string(buf))
}

func TestBuffer(t *testing.T) {
	cfg := NewConfig()
	err := cfg.Read("../config.json")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Configuration read...")

	rt := &Runtime{}
	rt.Locations = cfg.Locations
	rt.Location = cfg.Locations[0]
	rt.QueryHourly()

	buf, err := json.MarshalIndent(&rt.Location.WeatherHourly, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("hourly", string(buf))

	const pattern = "../www/weather.html"
	templ, err := template.ParseGlob(pattern)
	if err != nil {
		log.Fatalln("ParseGlob", pattern, err)
	}
	b := bytes.Buffer{}
	tmp := templ.Lookup("weather.clock")
	if tmp == nil {
		t.Fatal("Lookup weather.clock nil")
	}
	tmp.Execute(&b, rt)
	t.Log(b.String())
}
