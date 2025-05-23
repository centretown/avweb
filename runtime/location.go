package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
)

type CurrentItem struct {
	ID          string
	Title       string
	Description string
	Klass       string
	Min         float64
	Max         float64
	ScaleMin    float64
	ScaleMax    float64
	Icon        string
	Color       string
	Units       string
	Chart       string
	Selected    bool
	Value       float64
}

type LocationItem struct {
	CurrentItem
	Values []float64
}

type LocationProperties struct {
	Index int
	Items []*LocationItem
	Code  []int32
}

type CurrentProperties struct {
	Index int
	Items []*CurrentItem
	Code  int32
}

type Location struct {
	ID                uint64              `json:"-" db:"ID"`
	City              string              `json:"city" db:"City"`
	Latitude          float64             `json:"latitude" db:"Latitude"`
	Longitude         float64             `json:"longitude" db:"Longitude"`
	Zone              string              `json:"zone" db:"Zone"`
	WeatherDaily      *WeatherDaily       `json:"-"`
	DailyProperties   *LocationProperties `json:"-"`
	WeatherHourly     *WeatherHourly      `json:"-"`
	HourlyProperties  *LocationProperties `json:"-"`
	WeatherCurrent    *WeatherCurrent     `json:"-"`
	CurrentProperties *CurrentProperties  `json:"-"`
	History           []*Current          `json:"-"`
	HistoryProperties *CurrentProperties  `json:"-"`
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

func (loc *Location) QueryCurrent(db *sqlx.DB) (err error) {
	var (
		trailer = "&current=temperature_2m,precipitation,relative_humidity_2m,apparent_temperature,is_day,weather_code,wind_speed_10m,wind_direction_10m,wind_gusts_10m,rain,showers,cloud_cover,pressure_msl,surface_pressure,snowfall"
	)
	q := fmt.Sprintf(format, header, loc.Latitude, loc.Longitude, loc.Zone, trailer)
	loc.WeatherCurrent, err = GetWeatherCurrent(q)
	if err != nil {
		log.Println("QueryCurrent", err)
		return
	}
	err = InsertHistory(db, loc.ID, loc.WeatherCurrent.Current)
	return
}

var currentKeys = []string{
	"temperature", "feelslike",
	"precipitation", "rain",
	"shower", "snow",
	"cloud", "humidity",
	"windspeed", "windgusts",
	"surface", "pressure",
}

var hourlyKeys = []string{
	"temperature", "feelslike",
	"precipitation", "probability",
	"windspeed", "windgusts",
	"pressure", "humidity",
}

var dailyKeys = []string{
	"temperature-high", "temperature-low",
	"precipitation", "probability",
	"windspeed", "windgusts",
	"daylight", "sunshine",
}

func (loc *Location) BuildHistoryProperties(index int) *CurrentProperties {
	p := &CurrentProperties{}
	AttributesCurrent(index, p, loc.History[index], &loc.WeatherCurrent.CurrentUnits)
	loc.HistoryProperties = p
	return p
}

func (loc *Location) BuildCurrentProperties(index int) {
	p := &CurrentProperties{}
	AttributesCurrent(index, p, loc.WeatherCurrent.Current, &loc.WeatherCurrent.CurrentUnits)
	loc.CurrentProperties = p
}

func AttributesCurrent(index int, p *CurrentProperties, values *Current, units *CurrentUnits) {
	p.Index = index
	p.Items = make([]*CurrentItem, len(currentKeys))
	p.Code = values.Code
	for i, key := range currentKeys {
		item := &CurrentItem{}
		p.Items[i] = item

		item.ID = fmt.Sprintf("%s%d", key, index)
		item.Klass = key

		attr := Attributes[key]
		attr.ToItem(item)
		switch key {
		case "temperature":
			item.Value = values.Temperature
			item.Units = units.Temperature
		case "feelslike":
			item.Value = values.FeelsLike
			item.Units = units.FeelsLike
		case "precipitation":
			item.Value = values.Precipitation
			item.Units = units.Precipitation
		case "windspeed":
			item.Value = values.WindSpeed
			item.Units = units.WindSpeed
		case "windgusts":
			item.Value = values.WindGusts
			item.Units = units.WindGusts
		case "pressure":
			item.Value = values.PressureMSL
			item.Units = units.PressureMSL
		case "surface":
			item.Value = values.SurfacePressure
			item.Units = units.SurfacePressure
		case "humidity":
			item.Value = values.Humidity
			item.Units = units.Humidity
		case "rain":
			item.Value = values.Rain
			item.Units = units.Rain
		case "shower":
			item.Value = values.Showers
			item.Units = units.Showers
		case "snow":
			item.Value = values.Snowfall
			item.Units = units.Snowfall
		case "cloud":
			item.Value = values.CloudCover
			item.Units = units.CloudCover
		}
		item.Max = item.Value
		item.Min = 0
	}
}

func (loc *Location) BuildDailyProperties(index int) {
	p := &LocationProperties{}
	loc.DailyProperties = p
	p.Index = index
	p.Items = make([]*LocationItem, len(dailyKeys))
	p.Code = loc.WeatherDaily.Daily.Code
	limits := make(map[string]*Limits)
	for i, key := range dailyKeys {
		item := &LocationItem{}
		p.Items[i] = item

		item.ID = fmt.Sprintf("%s%d", key, index)
		item.Klass = key

		attr := Attributes[key]
		attr.ToItem(&item.CurrentItem)
		values := loc.WeatherDaily.Daily
		units := loc.WeatherDaily.DailyUnits

		switch key {
		case "temperature-high":
			item.Values = values.High
			item.Units = units.High
		case "temperature-low":
			item.Values = values.Low
			item.Units = units.Low
		case "precipitation":
			item.Values = values.Precipitation
			item.Units = units.Precipitation
		case "probability":
			item.Values = values.Probability
			item.Units = units.Probability
		case "windspeed":
			item.Values = values.WindSpeed
			item.Units = units.WindSpeed
		case "windgusts":
			item.Values = values.WindGusts
			item.Units = units.WindGusts
		case "daylight":
			item.Values = values.Daylight
			item.Units = units.Daylight
		case "sunshine":
			item.Values = values.Sunshine
			item.Units = units.Sunshine
		}

		mnx := loc.WeatherDaily.MinMax(item.Values)
		item.Max = mnx.Max
		item.Min = mnx.Min
		p.BuildScale(limits, &mnx, item.Units)
	}

	p.Scale(limits)
}

func (loc *Location) BuildHourlyProperties(index int) {
	p := &LocationProperties{}
	loc.HourlyProperties = p
	p.Index = index
	p.Items = make([]*LocationItem, len(hourlyKeys))
	p.Code = loc.WeatherHourly.Hourly.Code
	limits := make(map[string]*Limits)

	values := loc.WeatherHourly.Hourly
	units := loc.WeatherHourly.HourlyUnits

	for i, key := range hourlyKeys {
		item := &LocationItem{}
		p.Items[i] = item

		item.ID = fmt.Sprintf("%s%d", key, index)
		item.Klass = key

		attr := Attributes[key]
		attr.ToItem(&item.CurrentItem)

		switch key {
		case "temperature":
			item.Values = values.Temperature
			item.Units = units.Temperature
		case "feelslike":
			item.Values = values.FeelsLike
			item.Units = units.FeelsLike
		case "precipitation":
			item.Values = values.Precipitation
			item.Units = units.Precipitation
		case "probability":
			item.Values = values.Probability
			item.Units = units.Probability
		case "windspeed":
			item.Values = values.WindSpeed
			item.Units = units.WindSpeed
		case "windgusts":
			item.Values = values.WindGusts
			item.Units = units.WindGusts
		case "pressure":
			item.Values = values.Pressure
			item.Units = units.Pressure
		case "humidity":
			item.Values = values.Humidity
			item.Units = units.Humidity
		}

		mnx := loc.WeatherHourly.MinMax(item.Values)
		item.Max = mnx.Max
		item.Min = mnx.Min
		p.BuildScale(limits, &mnx, item.Units)
	}

	p.Scale(limits)
}

func (p *LocationProperties) BuildScale(limits map[string]*Limits, mnx *Limits, units string) {
	lim, ok := limits[units]
	if !ok {
		limits[units] = mnx
	} else {
		// log.Println(units, lim.Min, lim.Max)
		if lim.Min > mnx.Min {
			lim.Min = mnx.Min
		}
		if lim.Max < mnx.Max {
			lim.Max = mnx.Max
		}
	}
}

func (p *LocationProperties) Scale(limits map[string]*Limits) {
	for _, item := range p.Items {
		lim, ok := limits[item.Units]
		if !ok {
			item.ScaleMax = 100.0
			item.ScaleMin = 0.0
		} else {
			if item.Units == "%" {
				item.ScaleMax = 100.0
				item.ScaleMin = 0.0
			} else if item.Units == "s" {
				item.ScaleMax = item.Max
				item.ScaleMin = item.Min
			} else {
				item.ScaleMax = lim.Max
				item.ScaleMin = lim.Min
			}
		}
	}
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
