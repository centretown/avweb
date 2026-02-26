package runtime

import (
	"log"
	"time"
)

type Limits struct {
	Min float64
	Max float64
}

type WeatherCommon struct {
	Latitude             float64   `json:"latitude"`
	Longitude            float64   `json:"longitude"`
	GenerationtimeMs     float64   `json:"generationtime_ms"`
	UtcOffsetSeconds     float64   `json:"utc_offset_seconds"`
	Timezone             string    `json:"timezone"`
	TimezoneAbbreviation string    `json:"timezone_abbreviation"`
	Elevation            float64   `json:"elevation"`
	UpdateTime           time.Time `json:"-"`
}

type WeatherHourly struct {
	WeatherCommon
	HourlyUnits HourlyUnits `json:"hourly_units"`
	Hourly      *Hourly     `json:"hourly"`
}

type WeatherDaily struct {
	WeatherCommon
	DailyUnits DailyUnits `json:"daily_units"`
	Daily      *Daily     `json:"daily"`
}

type WeatherCurrent struct {
	WeatherCommon
	CurrentUnits CurrentUnits `json:"current_units"`
	Current      *Current     `json:"current"`
}

type CurrentUnits struct {
	Time            string `json:"time"`
	Temperature     string `json:"temperature_2m"`
	Precipitation   string `json:"precipitation"`
	Humidity        string `json:"relative_humidity_2m"`
	FeelsLike       string `json:"apparent_temperature"`
	WindSpeed       string `json:"wind_speed_10m"`
	WindDirection   string `json:"wind_direction_10m"`
	WindGusts       string `json:"wind_gusts_10m"`
	Rain            string `json:"rain"`
	Showers         string `json:"showers"`
	Snowfall        string `json:"snowfall"`
	CloudCover      string `json:"cloud_cover"`
	PressureMSL     string `json:"pressure_msl"`
	SurfacePressure string `json:"surface_pressure"`
	Code            string `json:"weather_code"`
	Interval        string `json:"interval"`
	IsDay           string `json:"is_day"`
}

type Current struct {
	LocationID      uint64  `json:"-" db:"LocationID"`
	Time            string  `json:"time" db:"Time"`
	Temperature     float64 `json:"temperature_2m" db:"Temperature"`
	Precipitation   float64 `json:"precipitation" db:"Precipitation"`
	Humidity        float64 `json:"relative_humidity_2m" db:"Humidity"`
	FeelsLike       float64 `json:"apparent_temperature" db:"FeelsLike"`
	WindSpeed       float64 `json:"wind_speed_10m" db:"WindSpeed"`
	WindDirection   float64 `json:"wind_direction_10m" db:"WindDirection"`
	WindGusts       float64 `json:"wind_gusts_10m" db:"WindGusts"`
	Rain            float64 `json:"rain" db:"Rain"`
	Showers         float64 `json:"showers" db:"Showers"`
	Snowfall        float64 `json:"snowfall" db:"Snowfall"`
	CloudCover      float64 `json:"cloud_cover" db:"CloudCover"`
	PressureMSL     float64 `json:"pressure_msl" db:"PressureMSL"`
	SurfacePressure float64 `json:"surface_pressure" db:"SurfacePressure"`
	Code            int32   `json:"weather_code" db:"Code"`
	Interval        int32   `json:"interval" db:"Interval"`
	IsDay           float32 `json:"is_day" db:"IsDay"`
}

const (
	HISTORY_INTERVALS       = 24
	HISTORY_MINUTE_INTERVAL = 15
)

type CurrentValues struct {
	Time            [HISTORY_INTERVALS]string
	Temperature     [HISTORY_INTERVALS]float64
	Precipitation   [HISTORY_INTERVALS]float64
	Humidity        [HISTORY_INTERVALS]float64
	FeelsLike       [HISTORY_INTERVALS]float64
	WindSpeed       [HISTORY_INTERVALS]float64
	WindDirection   [HISTORY_INTERVALS]float64
	WindGusts       [HISTORY_INTERVALS]float64
	Rain            [HISTORY_INTERVALS]float64
	Showers         [HISTORY_INTERVALS]float64
	Snowfall        [HISTORY_INTERVALS]float64
	CloudCover      [HISTORY_INTERVALS]float64
	PressureMSL     [HISTORY_INTERVALS]float64
	SurfacePressure [HISTORY_INTERVALS]float64
	Code            [HISTORY_INTERVALS]int32
	Interval        [HISTORY_INTERVALS]int32
	IsDay           [HISTORY_INTERVALS]int8
}

type HourlyUnits struct {
	Time          string `json:"time"`
	Temperature   string `json:"temperature_2m"`
	FeelsLike     string `json:"apparent_temperature"`
	Probability   string `json:"precipitation_probability"`
	Precipitation string `json:"precipitation"`
	WindSpeed     string `json:"wind_speed_10m"`
	WindDirection string `json:"wind_direction_10m"`
	WindGusts     string `json:"wind_gusts_10m"`
	Humidity      string `json:"relative_humidity_2m"`
	Pressure      string `json:"surface_pressure"`
	Code          string `json:"weather_code"`
}

type Hourly struct {
	Time          []string  `json:"time"`
	Temperature   []float64 `json:"temperature_2m"`
	FeelsLike     []float64 `json:"apparent_temperature"`
	Probability   []float64 `json:"precipitation_probability"`
	Precipitation []float64 `json:"precipitation"`
	WindSpeed     []float64 `json:"wind_speed_10m"`
	WindDirection []float64 `json:"wind_direction_10m"`
	WindGusts     []float64 `json:"wind_gusts_10m"`
	Humidity      []float64 `json:"relative_humidity_2m"`
	Pressure      []float64 `json:"surface_pressure"`
	Code          []int32   `json:"weather_code"`
}

type DailyUnits struct {
	Time          string `json:"time"`
	Sunrise       string `json:"sunrise"`
	Sunset        string `json:"sunset"`
	High          string `json:"temperature_2m_max"`
	Low           string `json:"temperature_2m_min"`
	Daylight      string `json:"daylight_duration"`
	Sunshine      string `json:"sunshine_duration"`
	Precipitation string `json:"precipitation_sum"`
	Probability   string `json:"precipitation_probability_max"`
	WindSpeed     string `json:"wind_speed_10m_max"`
	WindDirection string `json:"wind_direction_10m_dominant"`
	WindGusts     string `json:"wind_gusts_10m_max"`
	Code          string `json:"weather_code"`
}

type Daily struct {
	Time          []string  `json:"time"`
	Sunrise       []string  `json:"sunrise"`
	Sunset        []string  `json:"sunset"`
	High          []float64 `json:"temperature_2m_max"`
	Low           []float64 `json:"temperature_2m_min"`
	Daylight      []float64 `json:"daylight_duration"`
	Sunshine      []float64 `json:"sunshine_duration"`
	Precipitation []float64 `json:"precipitation_sum"`
	Probability   []float64 `json:"precipitation_probability_max"`
	WindSpeed     []float64 `json:"wind_speed_10m_max"`
	WindDirection []float64 `json:"wind_direction_10m_dominant"`
	WindGusts     []float64 `json:"wind_gusts_10m_max"`
	Code          []int32   `json:"weather_code"`
}

func (c *WeatherCommon) LogCommon() {
	log.Printf("latitude: %f longitude: %f \ngenerationtime_ms: %f utc_offset_seconds: %f \ntimezone: %s (%s) \nelevation %f\n",
		c.Latitude,
		c.Longitude,
		c.GenerationtimeMs,
		c.UtcOffsetSeconds,
		c.Timezone,
		c.TimezoneAbbreviation,
		c.Elevation)
}

func (w *WeatherCommon) FormatHour(hour time.Time) string {
	return hour.Format("3PM")
}

func (w *WeatherCommon) WeatherCode(codes int32) (wcode *WeatherCode) {
	return WeatherCodes[codes]
}

func (w *WeatherCommon) MinMax(args ...[]float64) (limits Limits) {
	limits.Min = 65000
	limits.Max = -65000
	for _, values := range args {
		for _, value := range values {
			if limits.Max < value {
				limits.Max = value
			}
			if limits.Min > value {
				limits.Min = value
			}
		}
	}
	return limits
}

func (w *WeatherCommon) FormatTimeCurrent(times []string) (intervals []string) {
	for index, timestr := range times {
		if index%4 != 0 {
			continue
		}
		t, err := time.Parse("2006-01-02T15:04", timestr)
		if err != nil {
			log.Printf("FormatTime: %v\n", err)
		}
		intervals = append(intervals, t.Format("3:04PM"))
	}

	return
}

func (w *WeatherCommon) FormatTime(timeStr string) string {
	t, err := time.Parse("2006-01-02T15:04", timeStr)
	if err != nil {
		log.Printf("FormatTime: %v\n", err)
	}
	return t.Format("Monday, Jan 2, 2006 at 3:04pm")
}

func (w *WeatherHourly) Hours() (hours []time.Time) {
	hours = make([]time.Time, 0, len(w.Hourly.Time)/4)
	for i := range w.Hourly.Time {
		if i%4 == 0 {
			t, _ := time.Parse("2006-01-02T15:04", w.Hourly.Time[i])
			hours = append(hours, t)
		}
	}
	return
}

func (w *WeatherDaily) FormatDay(index int) string {
	t, _ := time.Parse("2006-01-02", w.Daily.Time[index])
	return t.Format("Monday, Jan 2, 2006")
}

func (w *WeatherDaily) FormatDayShort(index int) string {
	t, _ := time.Parse("2006-01-02", w.Daily.Time[index])
	return t.Format("Mon")
}

type WeatherEffects struct {
	Token     string
	Class     string
	Intensity int
}

type WeatherCode struct {
	Code   int
	Color  string
	Icon   string
	Tokens []string
}

var WeatherCodes = map[int32]*WeatherCode{
	0:  {Code: 0, Color: "yellow", Icon: "clear_day", Tokens: []string{"clear", "sky"}},
	1:  {Code: 1, Color: "lightyellow", Icon: "clear_day", Tokens: []string{"mainly", "clear"}},
	2:  {Code: 2, Color: "lightblue", Icon: "partly_cloudy_day", Tokens: []string{"partly", "cloudy"}},
	3:  {Code: 3, Color: "darkgray", Icon: "cloud", Tokens: []string{"overcast", ""}},
	45: {Code: 45, Color: "dimgray", Icon: "foggy", Tokens: []string{"fog", ""}},
	48: {Code: 48, Color: "gray", Icon: "mist", Tokens: []string{"rime", "fog"}},
	51: {Code: 51, Color: "dodgerblue", Icon: "rainy", Tokens: []string{"light", "drizzle"}},
	53: {Code: 53, Color: "dodgerblue", Icon: "rainy", Tokens: []string{"moderate", "drizzle"}},
	55: {Code: 55, Color: "dodgerblue", Icon: "rainy", Tokens: []string{"dense", "drizzle"}},
	56: {Code: 56, Color: "lightblue", Icon: "rainy_snow", Tokens: []string{"light", "freezing", "drizzle"}},
	57: {Code: 57, Color: "lightblue", Icon: "rainy_snow", Tokens: []string{"dense", "freezing", "drizzle"}},
	61: {Code: 61, Color: "royalblue", Icon: "rainy_light", Tokens: []string{"slight", "rain"}},
	63: {Code: 63, Color: "royalblue", Icon: "rainy_light", Tokens: []string{"moderate", "rain"}},
	65: {Code: 65, Color: "royalblue", Icon: "rainy_heavy", Tokens: []string{"heavy", "rain"}},
	66: {Code: 66, Color: "slateblue", Icon: "rainy_heavy", Tokens: []string{"light", "freezing", "rain"}},
	67: {Code: 67, Color: "slateblue", Icon: "rainy_heavy", Tokens: []string{"heavy", "freezing", "rain"}},
	71: {Code: 71, Color: "white", Icon: "snowing", Tokens: []string{"slight", "snow"}},
	73: {Code: 73, Color: "white", Icon: "snowing", Tokens: []string{"moderate", "snow"}},
	75: {Code: 75, Color: "white", Icon: "snowing_heavy", Tokens: []string{"heavy", "snow"}},
	77: {Code: 77, Color: "white", Icon: "snowing", Tokens: []string{"snow", "grains"}},
	80: {Code: 80, Color: "dodgerblue", Icon: "shower", Tokens: []string{"slight", "rain", "showers"}},
	81: {Code: 81, Color: "dodgerblue", Icon: "shower", Tokens: []string{"moderate", "rain", "showers"}},
	82: {Code: 82, Color: "red", Icon: "rainy_heavy", Tokens: []string{"violent", "rain", "showers"}},
	85: {Code: 85, Color: "whitesmoke", Icon: "snowing", Tokens: []string{"slight", "snow", "showers"}},
	86: {Code: 86, Color: "white", Icon: "snowing_heavy", Tokens: []string{"heavy", "snow", "showers"}},
	95: {Code: 95, Color: "red", Icon: "thunderstorm", Tokens: []string{"thunderstorm"}},
	96: {Code: 96, Color: "pink", Icon: "weather_hail", Tokens: []string{"slight", "hail", "thunderstorm"}},
	99: {Code: 99, Color: "red", Icon: "weather_hail", Tokens: []string{"heavy", "hail", "thunderstorm"}},
}
