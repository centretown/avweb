package runtime

import "fmt"

type Location struct {
	City          string
	Latitude      float64
	Longitude     float64
	Zone          string
	WeatherDaily  WeatherDaily
	WeatherHourly WeatherHourly
}

var (
	format = "%s?latitude=%.2f&longitude=%.2f&timezone=%s%s"
	header = "https://api.open-meteo.com/v1/forecast"
)

func (loc *Location) QueryDaily() (err error) {
	var (
		trailer = "&daily=sunrise,sunset,temperature_2m_max,temperature_2m_min&hourly=temperature_2m,apparent_temperature,precipitation_probability,precipitation,weather_code,wind_speed_10m"
	)
	q := fmt.Sprintf(format, header, loc.Latitude, loc.Longitude, loc.Zone, trailer)
	return loc.WeatherDaily.Get(q)
}

func (loc *Location) QueryHourly() (err error) {
	var (
		trailer = "&hourly=temperature_2m,apparent_temperature,precipitation_probability,precipitation,weather_code,wind_speed_10m"
	)
	q := fmt.Sprintf(format, header, loc.Latitude, loc.Longitude, loc.Zone, trailer)
	return loc.WeatherHourly.Get(q)
}
