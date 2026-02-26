package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	weatherFormat  = "%s?latitude=%.2f&longitude=%.2f&timezone=%s%s&models=gem_seamless"
	weatherHeader  = "https://api.open-meteo.com/v1/forecast"
	dailyTrailer   = "&daily=sunrise,sunset,temperature_2m_max,temperature_2m_min,daylight_duration,sunshine_duration,precipitation_sum,precipitation_probability_max,weather_code,wind_speed_10m_max,wind_direction_10m_dominant,wind_gusts_10m_max"
	hourlyTrailer  = "&hourly=temperature_2m,apparent_temperature,precipitation_probability,precipitation,weather_code,wind_speed_10m,wind_direction_10m,wind_gusts_10m,relative_humidity_2m,surface_pressure,&forecast_hours=24"
	currentTrailer = "&current=temperature_2m,precipitation,relative_humidity_2m,apparent_temperature,is_day,weather_code,wind_speed_10m,wind_direction_10m,wind_gusts_10m,rain,showers,cloud_cover,pressure_msl,surface_pressure,snowfall"
)

func FirstTicker() (ticker time.Duration) {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(),
		now.Minute(), 0, 0, now.Location())
	minute := 15 - next.Minute()%15
	if minute == 0 {
		minute = 15
	}
	next = next.Add(time.Duration(minute) * time.Minute)
	if next.Compare(now) < 0 {
		log.Fatal(minute, now, next)
	}
	ticker = next.Sub(now)
	log.Printf("ticker=%v, minute=%v now=%v next=%v\n", ticker, minute, now, next)
	return
}

func (rt *Runtime) Monitor() {
	var (
		now         time.Time
		resetTicker = true
	)

	for {
		now = <-rt.ticker.C
		if resetTicker {
			rt.ticker.Reset(time.Minute * 15)
			resetTicker = false
		}
		rt.QueryCurrent()
		if now.Minute() == 0 {
			rt.QueryHourly()
			if now.Hour()%4 == 0 {
				rt.QueryDaily()
			}
		}
		rt.BroadcastTemperature()
		time.Sleep(time.Second)
	}
}

func (rt *Runtime) QueryDaily() {
	log.Println("Retrieving daily weather forecast...")
	for _, location := range rt.Locations {
		daily := &WeatherDaily{}
		query := fmt.Sprintf(weatherFormat, weatherHeader, location.Latitude, location.Longitude, location.Zone, dailyTrailer)
		err := queryAndDecode(query, daily)
		if err != nil {
			log.Printf("QueryDaily: queryAndDecode %v", err)
			continue
		}
		location.WeatherDaily = daily
		location.WeatherDaily.UpdateTime = time.Now()
		location.BuildDailyProperties()
	}
}

type LocationData struct {
	Index    int
	Location *Location
}

func (rt *Runtime) QueryHourly() {
	log.Println("Retrieving hourly weather forecast...")
	for _, location := range rt.Locations {
		query := fmt.Sprintf(weatherFormat, weatherHeader, location.Latitude, location.Longitude, location.Zone, hourlyTrailer)
		hourly := &WeatherHourly{}
		err := queryAndDecode(query, hourly)
		if err != nil {
			log.Printf("QueryHourly queryAndDecode: %v", err)
			continue
		}
		location.WeatherHourly = hourly
		location.WeatherHourly.UpdateTime = time.Now()
		location.BuildHourlyProperties()
	}
}

func (rt *Runtime) QueryCurrent() {
	log.Println("Retrieving current weather conditions...")
	for _, location := range rt.Locations {
		current := &WeatherCurrent{}
		query := fmt.Sprintf(weatherFormat, weatherHeader, location.Latitude, location.Longitude, location.Zone, currentTrailer)
		err := queryAndDecode(query, current)
		if err != nil {
			continue
		}
		location.WeatherCurrent = current
		err = InsertHistory(rt.db, location.ID, location.WeatherCurrent.Current)
		if err != nil {
			continue
		}
		location.WeatherCurrent.UpdateTime = time.Now()
	}

	err := rt.LoadHistory()
	if err != nil {
		log.Printf("Weather current load history error: %v", err)
	}
}

func queryAndDecode(query string, w any) (err error) {
	var resp *http.Response

	for attempt := range 3 {
		resp, err = http.Get(query)
		if err != nil {
			log.Printf("query current attempt: %v, error: %v", attempt, err)
			time.Sleep(time.Second)
			continue
		}

		defer resp.Body.Close()
		err = readAndDecode(resp.Body, w)
		if err != nil {
			log.Printf("decode current attempt: %v, error: %v", attempt, err)
			time.Sleep(time.Second)
			continue
		}

		return
	}

	return
}

func readAndDecode(r io.ReadCloser, w any) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %v", err)
	}
	err = json.Unmarshal(buf, w)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %v\n%v", err, string(buf))
	}
	return nil
}
