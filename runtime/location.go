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

// "https://api.open-meteo.com/v1/forecast?latitude=45.42&longitude=-75.7&daily=sunrise,sunset,temperature_2m_max,temperature_2m_min,daylight_duration,sunshine_duration,precipitation_sum,weather_code,uv_index_max&hourly=temperature_2m,apparent_temperature,precipitation_probability,precipitation,weather_code,wind_speed_10m&timezone=America%2FNew_York"
// "https://api.open-meteo.com/v1/forecast?latitude=45.42&longitude=-75.7&daily=sunrise,sunset,temperature_2m_max,temperature_2m_min,daylight_duration,sunshine_duration,precipitation_sum,weather_code,uv_index_max&hourly=temperature_2m,apparent_temperature,precipitation_probability,precipitation,weather_code,wind_speed_10m&timezone=America%2FNew_York"

var (
	format = "%s?latitude=%.2f&longitude=%.2f&timezone=%s%s"
	header = "https://api.open-meteo.com/v1/forecast"
)

func (loc *Location) QueryDaily() (err error) {
	var (
		// trailer = "&daily=sunrise,sunset,temperature_2m_max,temperature_2m_min"
		trailer = "&daily=sunrise,sunset,temperature_2m_max,temperature_2m_min,daylight_duration,sunshine_duration,precipitation_sum,weather_code,uv_index_max"
	)
	q := fmt.Sprintf(format, header, loc.Latitude, loc.Longitude, loc.Zone, trailer)
	return loc.WeatherDaily.Get(q)
}

func (loc *Location) QueryHourly() (err error) {
	var (
		trailer = "&hourly=temperature_2m,apparent_temperature,precipitation_probability,precipitation,weather_code,wind_speed_10m&forecast_hours=25&past_hours=12"
	)
	q := fmt.Sprintf(format, header, loc.Latitude, loc.Longitude, loc.Zone, trailer)
	return loc.WeatherHourly.Get(q)
}

func (loc *Location) QueryMinutely() (err error) {
	var (
		trailer = "&minutely_15=temperature_2m,apparent_temperature,precipitation_probability,precipitation,weather_code,wind_speed_10m&forecast_hours=4&past_hours=2"
	)
	q := fmt.Sprintf(format, header, loc.Latitude, loc.Longitude, loc.Zone, trailer)
	return loc.WeatherHourly.Get(q)
}
