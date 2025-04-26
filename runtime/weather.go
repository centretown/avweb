package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type WeatherCommon struct {
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	GenerationtimeMs     float64 `json:"generationtime_ms"`
	UtcOffsetSeconds     int64   `json:"utc_offset_seconds"`
	Timezone             string  `json:"timezone"`
	TimezoneAbbreviation string  `json:"timezone_abbreviation"`
	Elevation            float64 `json:"elevation"`
}

type WeatherHourly struct {
	WeatherCommon
	HourlyUnits HourlyUnits `json:"hourly_units"`
	Hourly      Hourly      `json:"hourly"`
}

type WeatherDaily struct {
	WeatherCommon
	DailyUnits DailyUnits `json:"daily_units"`
	Daily      Daily      `json:"daily"`
}

type HourlyUnits struct {
	Time                     string `json:"time"`
	Temperature2M            string `json:"temperature_2m"`
	ApparentTemperature      string `json:"apparent_temperature"`
	PrecipitationProbability string `json:"precipitation_probability"`
	Precipitation            string `json:"precipitation"`
	WeatherCode              string `json:"weather_code"`
	WindSpeed10M             string `json:"wind_speed_10m"`
}

type Hourly struct {
	Time                     []string  `json:"time"`
	Temperature              []float32 `json:"temperature_2m"`
	ApparentTemperature      []float32 `json:"apparent_temperature"`
	PrecipitationProbability []int32   `json:"precipitation_probability"`
	Precipitation            []float32 `json:"precipitation"`
	WeatherCode              []int32   `json:"weather_code"`
	WindSpeed                []float32 `json:"wind_speed_10m"`
}

type DailyUnits struct {
	Time    string `json:"time"`
	Sunrise string `json:"sunrise"`
	Sunset  string `json:"sunset"`
	High    string `json:"temperature_2m_max"`
	Low     string `json:"temperature_2m_min"`
}

type Daily struct {
	Time    []string  `json:"time"`
	Sunrise []string  `json:"sunrise"`
	Sunset  []string  `json:"sunset"`
	High    []float64 `json:"temperature_2m_max"`
	Low     []float64 `json:"temperature_2m_min"`
}

func (w *WeatherDaily) FormatDay(index int) string {
	t, _ := time.Parse("2006-01-02", w.Daily.Time[index])
	return t.Format("Monday")
}

func (w *WeatherDaily) FormatHigh(index int) string {
	return fmt.Sprintf("%3.1f %s", w.Daily.High[index], w.DailyUnits.High)
}
func (w *WeatherDaily) FormatLow(index int) string {
	return fmt.Sprintf("%3.1f %s", w.Daily.Low[index], w.DailyUnits.Low)
}
func (w *WeatherDaily) FormatSunset(index int) string {
	t, err := time.Parse("2006-01-02T15:04", w.Daily.Sunset[index])
	if err != nil {
		log.Printf("FormatSunset: %v\n", err)
	}
	return t.Format("3:04PM")
}
func (w *WeatherDaily) FormatSunrise(index int) string {
	t, err := time.Parse("2006-01-02T15:04", w.Daily.Sunrise[index])
	if err != nil {
		log.Printf("FormatSunrise: %v\n", err)
	}
	return t.Format("3:04PM")
}

const DailyQuery = "https://api.open-meteo.com/v1/forecast?latitude=45.42&longitude=-75.7&daily=sunrise,sunset,temperature_2m_max,temperature_2m_min&timezone=America%2FNew_York"
const HourlyQuery = "https://api.open-meteo.com/v1/forecast?latitude=45.42&longitude=-75.7&hourly=temperature_2m,apparent_temperature,precipitation_probability,precipitation,weather_code,wind_speed_10m&timezone=America%2FNew_York"

func (c *WeatherCommon) LogCommon() {
	log.Printf("latitude: %f longitude: %f \ngenerationtime_ms: %f utc_offset_seconds: %d \ntimezone: %s (%s) \nelevation %f\n",
		c.Latitude,
		c.Longitude,
		c.GenerationtimeMs,
		c.UtcOffsetSeconds,
		c.Timezone,
		c.TimezoneAbbreviation,
		c.Elevation)
}

func (w *WeatherDaily) Log() {
	w.LogCommon()
	for i := range w.Daily.Time {
		log.Printf("date: %s\tsunrise: %s\tsunset: %s\n",
			w.Daily.Time[i], w.Daily.Sunrise[i], w.Daily.Sunset[i])
	}
}

func (daily *WeatherDaily) Get(query string) (err error) {
	return GetWeather(query, daily)
}

func (hourly *WeatherHourly) Log() {
	hourly.LogCommon()
}

func (hourly *WeatherHourly) Get(query string) (err error) {
	return GetWeather(query, hourly)
}

func GetWeather(query string, w any) (err error) {
	var (
		resp *http.Response
	)

	resp, err = http.Get(query)
	if err != nil {
		err = fmt.Errorf("weather Get: %v", err)
		return
	}
	defer resp.Body.Close()
	err = LoadWeather(resp.Body, w)
	return
}

func LoadWeather(r io.Reader, w any) (err error) {
	var (
		buf []byte
	)

	buf, err = io.ReadAll(r)
	if err != nil {
		err = fmt.Errorf("LoadWeather ReadAll: %v", err)
		return
	}

	err = json.Unmarshal(buf, w)
	if err != nil {
		err = fmt.Errorf("LoadWeather Unmarshal: %v", err)
		return
	}
	return
}
