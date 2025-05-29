package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"

	"github.com/jmoiron/sqlx"
)

type CurrentProperty struct {
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

type LocationProperty struct {
	CurrentProperty
	Values []float64
}

type LocationProperties struct {
	Index int
	Items []*LocationProperty
	Code  []int32
	Time  []string
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
	CurrentProperties *LocationProperties `json:"-"`
}

const (
	format         = "%s?latitude=%.2f&longitude=%.2f&timezone=%s%s&models=gem_seamless"
	header         = "https://api.open-meteo.com/v1/forecast"
	dailyTrailer   = "&daily=sunrise,sunset,temperature_2m_max,temperature_2m_min,daylight_duration,sunshine_duration,precipitation_sum,precipitation_probability_max,weather_code,wind_speed_10m_max,wind_direction_10m_dominant,wind_gusts_10m_max"
	hourlytrailer  = "&hourly=temperature_2m,apparent_temperature,precipitation_probability,precipitation,weather_code,wind_speed_10m,wind_direction_10m,wind_gusts_10m,relative_humidity_2m,surface_pressure,&forecast_hours=24"
	currentTrailer = "&current=temperature_2m,precipitation,relative_humidity_2m,apparent_temperature,is_day,weather_code,wind_speed_10m,wind_direction_10m,wind_gusts_10m,rain,showers,cloud_cover,pressure_msl,surface_pressure,snowfall"
)

var (
	currentAttributes = []string{
		TEMPERATURE,
		FEELSLIKE,
		PRECIPITATION,
		RAIN,
		SHOWER,
		SNOW,
		CLOUD,
		HUMIDITY,
		WINDSPEED,
		WINDGUSTS,
		SURFACE,
		PRESSURE,
	}
	hourlyAttributes = []string{
		TEMPERATURE,
		FEELSLIKE,
		PRECIPITATION,
		PROBABILITY,
		WINDSPEED,
		WINDGUSTS,
		PRESSURE,
		HUMIDITY,
	}
	dailyAttributes = []string{
		TEMPERATURE_HIGH,
		TEMPERATURE_LOW,
		PRECIPITATION,
		PROBABILITY,
		WINDSPEED,
		WINDGUSTS,
		DAYLIGHT,
		SUNSHINE,
	}
)

func (loc *Location) QueryDaily() (err error) {
	q := fmt.Sprintf(format, header, loc.Latitude, loc.Longitude, loc.Zone, dailyTrailer)
	loc.WeatherDaily, err = GetWeatherDaily(q)
	return
}

func (loc *Location) QueryHourly() (err error) {
	q := fmt.Sprintf(format, header, loc.Latitude, loc.Longitude, loc.Zone, hourlytrailer)
	loc.WeatherHourly, err = GetWeatherHourly(q)
	return
}

func (loc *Location) QueryCurrent(db *sqlx.DB) (err error) {
	q := fmt.Sprintf(format, header, loc.Latitude, loc.Longitude, loc.Zone, currentTrailer)
	loc.WeatherCurrent, err = GetWeatherCurrent(q)
	if err != nil {
		log.Println("QueryCurrent", err)
		return
	}
	err = InsertHistory(db, loc.ID, loc.WeatherCurrent.Current)
	return
}

func (loc *Location) BuildCurrentProperties(history []*Current) {
	p := &LocationProperties{}
	loc.CurrentProperties = p
	p.Items = make([]*LocationProperty, len(currentAttributes))
	p.Code = make([]int32, len(history))
	p.Time = make([]string, len(history))
	limits := make(map[string]*Limits)
	for recno, rec := range history {
		p.Time[recno] = rec.Time
		p.Code[recno] = rec.Code
	}
	for i, key := range currentAttributes {
		item := &LocationProperty{}
		p.Items[i] = item
		attr := Attributes[key]
		attr.ToItem(&item.CurrentProperty)
		item.Values = make([]float64, len(history))
		item.Units = CurrentUnit(key, &loc.WeatherCurrent.CurrentUnits)
		item.Klass = key
		item.ID = fmt.Sprintf("%s%d", key, loc.ID)
		for recno, rec := range history {
			item.Values[recno] = CurrentValue(key, rec)
		}
		mnx := loc.WeatherCurrent.MinMax(item.Values)
		item.Max = mnx.Max
		item.Min = mnx.Min
		p.BuildScale(limits, &mnx, item.Units)
	}
	p.Scale(limits)
}

func CurrentValue(key string, values *Current) (value float64) {
	switch key {
	case TEMPERATURE:
		value = values.Temperature
	case FEELSLIKE:
		value = values.FeelsLike
	case PRECIPITATION:
		value = values.Precipitation
	case WINDSPEED:
		value = values.WindSpeed
	case WINDGUSTS:
		value = values.WindGusts
	case PRESSURE:
		value = values.PressureMSL
	case SURFACE:
		value = values.SurfacePressure
	case HUMIDITY:
		value = values.Humidity
	case RAIN:
		value = values.Rain
	case SHOWER:
		value = values.Showers
	case SNOW:
		value = values.Snowfall
	case CLOUD:
		value = values.CloudCover
	}
	return
}

func CurrentUnit(key string, units *CurrentUnits) (unit string) {
	switch key {
	case TEMPERATURE:
		unit = units.Temperature
	case FEELSLIKE:
		unit = units.FeelsLike
	case PRECIPITATION:
		unit = units.Precipitation
	case WINDSPEED:
		unit = units.WindSpeed
	case WINDGUSTS:
		unit = units.WindGusts
	case PRESSURE:
		unit = units.PressureMSL
	case SURFACE:
		unit = units.SurfacePressure
	case HUMIDITY:
		unit = units.Humidity
	case RAIN:
		unit = units.Rain
	case SHOWER:
		unit = units.Showers
	case SNOW:
		unit = units.Snowfall
	case CLOUD:
		unit = units.CloudCover
	}
	return
}

func (loc *Location) BuildDailyProperties() {
	p := &LocationProperties{}
	loc.DailyProperties = p
	p.Index = int(loc.ID)
	p.Items = make([]*LocationProperty, len(dailyAttributes))
	p.Code = loc.WeatherDaily.Daily.Code
	limits := make(map[string]*Limits)
	for i, key := range dailyAttributes {
		item := &LocationProperty{}
		p.Items[i] = item

		item.ID = fmt.Sprintf("%s%d", key, int(loc.ID))
		item.Klass = key

		attr := Attributes[key]
		attr.ToItem(&item.CurrentProperty)
		values := loc.WeatherDaily.Daily
		units := loc.WeatherDaily.DailyUnits

		switch key {
		case TEMPERATURE_HIGH:
			item.Values = values.High
			item.Units = units.High
		case TEMPERATURE_LOW:
			item.Values = values.Low
			item.Units = units.Low
		case PRECIPITATION:
			item.Values = values.Precipitation
			item.Units = units.Precipitation
		case PROBABILITY:
			item.Values = values.Probability
			item.Units = units.Probability
		case WINDSPEED:
			item.Values = values.WindSpeed
			item.Units = units.WindSpeed
		case WINDGUSTS:
			item.Values = values.WindGusts
			item.Units = units.WindGusts
		case DAYLIGHT:
			item.Values = make([]float64, len(values.Daylight))
			for i, seconds := range values.Daylight {
				item.Values[i] = math.Round(100*seconds/60/60) / 100
			}
			item.Units = "hr"
		case SUNSHINE:
			item.Values = make([]float64, len(values.Sunshine))
			for i, seconds := range values.Sunshine {
				item.Values[i] = math.Round(100*seconds/60/60) / 100
			}
			item.Units = "hr"
		}

		mnx := loc.WeatherDaily.MinMax(item.Values)
		item.Max = mnx.Max
		item.Min = mnx.Min
		p.BuildScale(limits, &mnx, item.Units)
	}

	p.Scale(limits)
}

func (loc *Location) BuildHourlyProperties() {
	props := &LocationProperties{}
	defer func() {
		loc.HourlyProperties = props
	}()
	props.Index = int(loc.ID)
	props.Items = make([]*LocationProperty, len(hourlyAttributes))
	props.Code = loc.WeatherHourly.Hourly.Code
	limits := make(map[string]*Limits)

	values := loc.WeatherHourly.Hourly
	units := loc.WeatherHourly.HourlyUnits

	for i, key := range hourlyAttributes {
		item := &LocationProperty{}
		props.Items[i] = item

		item.ID = fmt.Sprintf("%s%d", key, loc.ID)
		item.Klass = key

		attr := Attributes[key]
		attr.ToItem(&item.CurrentProperty)

		switch key {
		case TEMPERATURE:
			item.Values = values.Temperature
			item.Units = units.Temperature
		case FEELSLIKE:
			item.Values = values.FeelsLike
			item.Units = units.FeelsLike
		case PRECIPITATION:
			item.Values = values.Precipitation
			item.Units = units.Precipitation
		case PROBABILITY:
			item.Values = values.Probability
			item.Units = units.Probability
		case WINDSPEED:
			item.Values = values.WindSpeed
			item.Units = units.WindSpeed
		case WINDGUSTS:
			item.Values = values.WindGusts
			item.Units = units.WindGusts
		case PRESSURE:
			item.Values = values.Pressure
			item.Units = units.Pressure
		case HUMIDITY:
			item.Values = values.Humidity
			item.Units = units.Humidity
		}

		mnx := loc.WeatherHourly.MinMax(item.Values)
		item.Max = mnx.Max
		item.Min = mnx.Min
		props.BuildScale(limits, &mnx, item.Units)
	}

	props.Scale(limits)
	return
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
		switch {
		case !ok, item.Units == "%":
			item.ScaleMax = 100.0
			item.ScaleMin = 0.0
		case item.Units == "hr":
			item.ScaleMax = item.Max
			item.ScaleMin = item.Min
		default:
			item.ScaleMax = lim.Max
			item.ScaleMin = lim.Min
		}
	}
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
