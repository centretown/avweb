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
	Min                  float64 `json:"-"`
	Max                  float64 `json:"-"`
}

func (w *WeatherCommon) FormatHour(hour time.Time) string {
	return hour.Format("3PM")
}

func (w *WeatherCommon) MinMax(args ...[]float64) float64 {
	w.Min = 65000
	w.Max = -65000
	for _, values := range args {
		for _, value := range values {
			if w.Max < value {
				w.Max = value
			}
			if w.Min > value {
				w.Min = value
			}
		}
	}
	return w.Max
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
type WeatherMinutely struct {
	WeatherCommon
	MinutelyUnits MinutelyUnits `json:"minutely_units"`
	Minutely      Minutely      `json:"minutely"`
}

type MinutelyUnits struct {
	Time          string `json:"time"`
	Temperature   string `json:"temperature_2m"`
	FeelsLike     string `json:"apparent_temperature"`
	Probability   string `json:"precipitation_probability"`
	Precipitation string `json:"precipitation"`
	WindSpeed     string `json:"wind_speed_10m"`
	Code          string `json:"weather_code"`
}

type Minutely struct {
	Time          string  `json:"time"`
	Temperature   float32 `json:"temperature_2m"`
	FeelsLike     float32 `json:"apparent_temperature"`
	Probability   int32   `json:"precipitation_probability"`
	Precipitation float32 `json:"precipitation"`
	WindSpeed     float32 `json:"wind_speed_10m"`
	Code          int32   `json:"weather_code"`
}

type HourlyUnits struct {
	Time          string `json:"time"`
	Temperature   string `json:"temperature_2m"`
	FeelsLike     string `json:"apparent_temperature"`
	Probability   string `json:"precipitation_probability"`
	Precipitation string `json:"precipitation"`
	WindSpeed     string `json:"wind_speed_10m"`
	Code          string `json:"weather_code"`
}

type Hourly struct {
	Time          []string  `json:"time"`
	Temperature   []float64 `json:"temperature_2m"`
	FeelsLike     []float64 `json:"apparent_temperature"`
	Probability   []int32   `json:"precipitation_probability"`
	Precipitation []float64 `json:"precipitation"`
	WindSpeed     []float64 `json:"wind_speed_10m"`
	Code          []int32   `json:"weather_code"`
}

func (w *WeatherHourly) FormatTime(index int) string {
	t, err := time.Parse("2006-01-02T15:04", w.Hourly.Time[index])
	if err != nil {
		log.Printf("FormatTime: %v\n", err)
	}
	return t.Format("3:04PM")
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

func (w *WeatherCommon) Icons(codes []int32, div int) (icons []string) {
	icons = make([]string, 0, len(codes)/int(div))
	for i := range codes {
		if i%div == 0 {
			icons = append(icons, WeatherCodes[int(codes[i])].Icon)
		}
	}
	return
}

func (w *WeatherHourly) FormatTemperature(index int) string {
	return fmt.Sprintf("%6.1f", w.Hourly.Temperature[index])
}
func (w *WeatherHourly) FormatFeelsLike(index int) string {
	return fmt.Sprintf("%6.1f", w.Hourly.FeelsLike[index])
}
func (w *WeatherHourly) FormatPrecipitation(index int) string {
	return fmt.Sprintf("%6.2f", w.Hourly.Precipitation[index])
}
func (w *WeatherHourly) FormatProbability(index int) string {
	return fmt.Sprintf("%6d", w.Hourly.Probability[index])
}
func (w *WeatherHourly) FormatWindSpeed(index int) string {
	return fmt.Sprintf("%6.2f", w.Hourly.WindSpeed[index])
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
	Code          string `json:"weather_code"`
	UvIndex       string `json:"uv_index_max"`
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
	Code          []int32   `json:"weather_code"`
	UvIndex       []float64 `json:"uv_index_max"`
}

func (w *WeatherDaily) ReadingsHigh() string {
	return fmt.Sprintf("// var readings = [%v];", w.Daily.High)
}

func (w *WeatherDaily) FormatDay(index int) string {
	t, _ := time.Parse("2006-01-02", w.Daily.Time[index])
	return t.Format("Monday")
}

func (w *WeatherDaily) FormatHigh(index int) string {
	return fmt.Sprintf("%4.0f %s", w.Daily.High[index], w.DailyUnits.High)
}
func (w *WeatherDaily) FormatPrecipitation(index int) string {
	return fmt.Sprintf("%4.2f %s", w.Daily.Precipitation[index], w.DailyUnits.Precipitation)
}
func (w *WeatherDaily) FormatUvIndex(index int) string {
	return fmt.Sprintf("%4.2f %s", w.Daily.UvIndex[index], w.DailyUnits.UvIndex)
}
func (w *WeatherDaily) FormatLow(index int) string {
	return fmt.Sprintf("%4.0f %s", w.Daily.Low[index], w.DailyUnits.Low)
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

func (w *WeatherDaily) FormatDayShort(index int) string {
	t, _ := time.Parse("2006-01-02", w.Daily.Time[index])
	return t.Format("Mon")
}

func toHours(fsec float64) string {
	seconds := int(fsec)
	hours := seconds / 3600
	seconds -= hours * 3600
	minutes := seconds / 60
	seconds = seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func (w *WeatherDaily) FormatDaylight(index int) string {
	return toHours(w.Daily.Daylight[index])
}
func (w *WeatherDaily) FormatSunshine(index int) string {
	return toHours(w.Daily.Sunshine[index])
}

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

type WeatherEffects struct {
	Token     string
	Class     string
	Intensity int
}

type WeatherCode struct {
	Code   int
	Icon   string
	Tokens []string
}

var WeatherCodes = map[int]*WeatherCode{
	0:  {Code: 0, Icon: "clear_day", Tokens: []string{"clear", "sky"}},
	1:  {Code: 1, Icon: "clear_day", Tokens: []string{"mainly", "clear"}},
	2:  {Code: 2, Icon: "partly_cloudy_day", Tokens: []string{"partly", "cloudy"}},
	3:  {Code: 3, Icon: "cloud", Tokens: []string{"overcast", ""}},
	45: {Code: 45, Icon: "foggy", Tokens: []string{"fog", ""}},
	48: {Code: 48, Icon: "mist", Tokens: []string{"rime", "fog"}},
	51: {Code: 51, Icon: "rainy_light", Tokens: []string{"light", "drizzle"}},
	53: {Code: 53, Icon: "rainy_light", Tokens: []string{"moderate", "drizzle"}},
	55: {Code: 55, Icon: "rainy_light", Tokens: []string{"dense", "drizzle"}},
	56: {Code: 56, Icon: "rainy_snow", Tokens: []string{"light", "freezing", "drizzle"}},
	57: {Code: 57, Icon: "rainy_snow", Tokens: []string{"dense", "freezing", "drizzle"}},
	61: {Code: 61, Icon: "rainy_light", Tokens: []string{"slight", "rain"}},
	63: {Code: 63, Icon: "rainy", Tokens: []string{"moderate", "rain"}},
	65: {Code: 65, Icon: "rainy_heavy", Tokens: []string{"heavy", "rain"}},
	66: {Code: 66, Icon: "rainy_heavy", Tokens: []string{"light", "freezing", "rain"}},
	67: {Code: 67, Icon: "rainy_heavy", Tokens: []string{"heavy", "freezing", "rain"}},
	71: {Code: 71, Icon: "snowing", Tokens: []string{"slight", "snow"}},
	73: {Code: 73, Icon: "snowing", Tokens: []string{"moderate", "snow"}},
	75: {Code: 75, Icon: "snowing_heavy", Tokens: []string{"heavy", "snow"}},
	77: {Code: 77, Icon: "snowing", Tokens: []string{"snow", "grains"}},
	80: {Code: 80, Icon: "rainy_light", Tokens: []string{"slight", "rain", "showers"}},
	81: {Code: 81, Icon: "rainy", Tokens: []string{"moderate", "rain", "showers"}},
	82: {Code: 82, Icon: "rainy_heavy", Tokens: []string{"violent", "rain", "showers"}},
	85: {Code: 85, Icon: "snowing", Tokens: []string{"slight", "snow", "showers"}},
	86: {Code: 86, Icon: "snowing_heavy", Tokens: []string{"heavy", "snow", "showers"}},
	95: {Code: 95, Icon: "thunderstorm", Tokens: []string{"thunderstorm"}},
	96: {Code: 96, Icon: "weather_hail", Tokens: []string{"slight", "hail", "thunderstorm"}},
	99: {Code: 99, Icon: "weather_hail", Tokens: []string{"heavy", "hail", "thunderstorm"}},
}

/*
0	Clear sky
1, 2, 3	Mainly clear, partly cloudy, and overcast
45, 48	Fog and depositing rime fog
51, 53, 55	Drizzle: Light, moderate, and dense intensity
56, 57	Freezing Drizzle: Light and dense intensity
61, 63, 65	Rain: Slight, moderate and heavy intensity
66, 67	Freezing Rain: Light and heavy intensity
71, 73, 75	Snow fall: Slight, moderate, and heavy intensity
77	Snow grains
80, 81, 82	Rain showers: Slight, moderate, and violent
85, 86	Snow showers slight and heavy
95 *	Thunderstorm: Slight or moderate
96, 99 *	Thunderstorm with slight and heavy hail
*/
