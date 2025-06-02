package homeasst

import (
	"strings"

	"github.com/centretown/avweb/action"
)

type Weather struct {
	Entity[WeatherAttributes]
	Action *action.Action
}

func (home *HomeRuntime) NewWeather(action *action.Action) *Weather {
	wthr := &Weather{
		Action: action,
	}
	entity, ok := home.Entities["weather.forecast_home"]
	if ok {
		wthr.Copy(entity)
	}
	wthr.Action.Icon = wthr.Icon()
	return wthr
}

// func (wthr *Weather) FormatTime() string {
// 	return wthr.LastUpdated.Local().Format(ShortTime)
// }

var weatherIcons = [][]string{
	{"partly_cloudy_day", "partly"},
	{"cloud", "cloud"},
	{"sunny", "sunny"},
	{"rainy", "rainy"},
	{"thunderstorm", "thunder"},
	{"storm", "storm"},
	{"clear_day", "clear"},
}

func (wthr *Weather) Icon() (icon string) {
	for _, keys := range weatherIcons {
		icon = keys[0]
		for _, k := range keys {
			if strings.Contains(wthr.State, k) {
				return
			}
		}
	}
	icon = "routine"
	return
}

type WeatherProperties struct {
	Icon  string
	Label string
	Value any
	Units string
}

func (wthr *Weather) Properties() []WeatherProperties {
	return []WeatherProperties{
		{Label: "Temperature", Icon: "device_thermostat",
			Units: wthr.Attributes.TemperatureUnit,
			Value: wthr.Attributes.Temperature},
		{Label: "Humidity", Icon: "humidity_percentage",
			Units: "%",
			Value: wthr.Attributes.Humidity},
		{Label: "Wind Speed", Icon: "air",
			Units: wthr.Attributes.WindSpeedUnit,
			Value: wthr.Attributes.WindSpeed},
		{Label: "Wind Bearing", Icon: "explore",
			Units: "\u00b0",
			Value: wthr.Attributes.WindBearing},
		{Label: "Cloud Coverage", Icon: "cloud",
			Units: "%",
			Value: wthr.Attributes.CloudCoverage},
		{Label: "Pressure", Icon: "compare_arrows",
			Units: wthr.Attributes.PressureUnit,
			Value: wthr.Attributes.Pressure},
		{Label: "Dew Point", Icon: "dew_point",
			Units: wthr.Attributes.PrecipitationUnit,
			Value: wthr.Attributes.DewPoint},
	}
}
