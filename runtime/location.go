package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type LocationItem struct {
	ID     string
	Klass  string
	Min    float64
	Max    float64
	Values []float64
	Icon   string
	Color  string
	Units  string
}

type LocationProperties struct {
	Index  int
	Items  []*LocationItem
	Limits map[string]*Limits
}

type Location struct {
	City             string              `json:"city"`
	Latitude         float64             `json:"latitude"`
	Longitude        float64             `json:"longitude"`
	Zone             string              `json:"zone"`
	WeatherDaily     *WeatherDaily       `json:"-"`
	WeatherHourly    *WeatherHourly      `json:"-"`
	HourlyProperties *LocationProperties `json:"-"`
	WeatherCurrent   *WeatherCurrent     `json:"-"`
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

type Pair struct {
	Icon  string
	Color string
}

var weatherIcons = map[string]Pair{
	"temperature":   {Icon: "thermometer", Color: "rgba(255, 150, 0, 255)"},
	"feelslike":     {Icon: "airwave", Color: "rgba(255, 200, 150, 255)"},
	"precipitation": {Icon: "weather_mix", Color: "rgba(31, 144, 255, 255)"},
	"probability":   {Icon: "weather_mix", Color: "rgba(31, 144, 255, 255)"},
	"windspeed":     {Icon: "toys_fan", Color: "yellow"},
	"windgusts":     {Icon: "air", Color: "yellow"},
	"pressure":      {Icon: "compress", Color: "rgba(31, 255, 31, 255)"},
	"humidity":      {Icon: "humidity_mid", Color: "rgba(192, 63, 255, 255)"},
}

var hourlyKeys = []string{
	"temperature", "feelslike", "precipitation", "probability",
	"windspeed", "windgusts", "pressure", "humidity",
}

func (loc *Location) BuildHourlyProperties(index int) {
	p := &LocationProperties{}
	loc.HourlyProperties = p
	p.Index = index
	p.Items = make([]*LocationItem, len(hourlyKeys))
	p.Limits = make(map[string]*Limits)

	var mnx Limits
	for i, key := range hourlyKeys {
		item := &LocationItem{}
		p.Items[i] = item

		item.ID = fmt.Sprintf("%s%d", key, index)
		item.Klass = key
		item.Icon = weatherIcons[key].Icon
		item.Color = weatherIcons[key].Color

		hourly := loc.WeatherHourly.Hourly
		hourlyUnits := loc.WeatherHourly.HourlyUnits
		switch key {
		case "temperature":
			item.Values = hourly.Temperature
			item.Units = hourlyUnits.Temperature
		case "feelslike":
			item.Values = hourly.FeelsLike
			item.Units = hourlyUnits.FeelsLike
		case "precipitation":
			item.Values = hourly.Precipitation
			item.Units = hourlyUnits.Precipitation
		case "probability":
			item.Values = hourly.Probability
			item.Units = hourlyUnits.Probability
		case "windspeed":
			item.Values = hourly.WindSpeed
			item.Units = hourlyUnits.WindSpeed
		case "windgusts":
			item.Values = hourly.WindGusts
			item.Units = hourlyUnits.WindGusts
		case "pressure":
			item.Values = hourly.Pressure
			item.Units = hourlyUnits.Pressure
		case "humidity":
			item.Values = hourly.Humidity
			item.Units = hourlyUnits.Humidity
		}

		mnx = loc.WeatherHourly.MinMax(item.Values)
		lim, ok := p.Limits[item.Units]
		if !ok {
			log.Println(item.Units, mnx)
			p.Limits[item.Units] = &mnx
		} else {
			log.Println(item.Units, lim.Min, lim.Max)
			if lim.Min > mnx.Min {
				lim.Min = mnx.Min
			}
			if lim.Max < mnx.Max {
				lim.Max = mnx.Max
			}
		}
		item.Max = mnx.Max
		item.Min = mnx.Min
	}

	// for _, item := range p.Items {
	// 	lim, ok := p.Limits[item.Units]
	// 	if !ok {
	// 		log.Println(item.Units, "not found")
	// 		item.Max = 0.0
	// 		item.Min = 100.0
	// 	} else {
	// 		log.Println(item.Units, *lim)
	// 		item.Max = lim.Max
	// 		item.Min = lim.Min
	// 	}
	// }
}

func GetWeatherDaily(query string) (daily *WeatherDaily, err error) {
	daily = &WeatherDaily{}
	// daily.UpdateTime = time.Now()
	err = GetWeather(query, daily)
	return daily, err
}

func GetWeatherHourly(query string) (hourly *WeatherHourly, err error) {
	hourly = &WeatherHourly{}
	// hourly.UpdateTime = time.Now()
	err = GetWeather(query, hourly)
	return hourly, err
}

func GetWeatherCurrent(query string) (current *WeatherCurrent, err error) {
	current = &WeatherCurrent{}
	// current.UpdateTime = time.Now()
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
