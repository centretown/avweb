package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Location struct {
	City           string          `json:"city"`
	Latitude       float64         `json:"latitude"`
	Longitude      float64         `json:"longitude"`
	Zone           string          `json:"zone"`
	WeatherDaily   *WeatherDaily   `json:"-"`
	WeatherHourly  *WeatherHourly  `json:"-"`
	WeatherCurrent *WeatherCurrent `json:"-"`
}

// "https://api.open-meteo.com/v1/forecast?latitude=45.42&longitude=-75.7&daily=sunrise,sunset,temperature_2m_max,temperature_2m_min,daylight_duration,sunshine_duration,precipitation_sum,weather_code,uv_index_max&hourly=temperature_2m,apparent_temperature,precipitation_probability,precipitation,weather_code,wind_speed_10m&timezone=America%2FNew_York"
// "https://api.open-meteo.com/v1/forecast?latitude=45.42&longitude=-75.7&daily=sunrise,sunset,temperature_2m_max,temperature_2m_min,daylight_duration,sunshine_duration,precipitation_sum,weather_code,uv_index_max&hourly=temperature_2m,apparent_temperature,precipitation_probability,precipitation,weather_code,wind_speed_10m&timezone=America%2FNew_York"
var (
	format = "%s?latitude=%.2f&longitude=%.2f&timezone=%s%s&models=gem_seamless"
	header = "https://api.open-meteo.com/v1/forecast"
)

func (loc *Location) QueryDaily() (err error) {
	var (
		trailer = "&daily=sunrise,sunset,temperature_2m_max,temperature_2m_min,daylight_duration,sunshine_duration,precipitation_sum,precipitation_probability_max,weather_code,wind_speed_10m_max,wind_direction_10m_dominant,wind_gusts_10m_max"
	)
	q := fmt.Sprintf(format, header, loc.Latitude, loc.Longitude, loc.Zone, trailer)
	loc.WeatherDaily, err = GetWeatherDaily(q)
	return
}

func (loc *Location) QueryHourly() (err error) {
	var (
		trailer = "&hourly=temperature_2m,apparent_temperature,precipitation_probability,precipitation,weather_code,wind_speed_10m,wind_direction_10m,wind_gusts_10m,relative_humidity_2m,surface_pressure,&forecast_hours=24"
	)
	q := fmt.Sprintf(format, header, loc.Latitude, loc.Longitude, loc.Zone, trailer)
	loc.WeatherHourly, err = GetWeatherHourly(q)
	return
}

func (loc *Location) QueryCurrent() (err error) {
	var (
		trailer = "&current=temperature_2m,precipitation,relative_humidity_2m,apparent_temperature,is_day,weather_code,wind_speed_10m,wind_direction_10m,wind_gusts_10m,rain,showers,cloud_cover,pressure_msl,surface_pressure,snowfall"
	)
	q := fmt.Sprintf(format, header, loc.Latitude, loc.Longitude, loc.Zone, trailer)
	loc.WeatherCurrent, err = GetWeatherCurrent(q)
	return
}

func GetWeatherDaily(query string) (daily *WeatherDaily, err error) {
	daily = &WeatherDaily{}
	err = GetWeather(query, daily)
	return daily, err
}

func GetWeatherHourly(query string) (hourly *WeatherHourly, err error) {
	hourly = &WeatherHourly{}
	err = GetWeather(query, hourly)
	return hourly, err
}

func GetWeatherCurrent(query string) (current *WeatherCurrent, err error) {
	current = &WeatherCurrent{}
	err = GetWeather(query, current)
	return current, err
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
