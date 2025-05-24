package runtime

type WeatherAttributes struct {
	Icon        string
	Color       string
	Chart       string
	Title       string
	Description string
	Selected    bool
}

func (attr *WeatherAttributes) ToItem(item *CurrentItem) {
	item.Icon = attr.Icon
	item.Title = attr.Title
	item.Description = attr.Description
	item.Color = attr.Color
	item.Chart = attr.Chart
	item.Selected = attr.Selected
}

const (
	TEMPERATURE      = "temperature"
	FEELSLIKE        = "feelslike"
	PRECIPITATION    = "precipitation"
	RAIN             = "rain"
	SHOWER           = "shower"
	SNOW             = "snow"
	CLOUD            = "cloud"
	HUMIDITY         = "humidity"
	WINDSPEED        = "windspeed"
	WINDGUSTS        = "windgusts"
	SURFACE          = "surface"
	PRESSURE         = "pressure"
	PROBABILITY      = "probability"
	DAYLIGHT         = "daylight"
	SUNSHINE         = "sunshine"
	TEMPERATURE_HIGH = "temperature-high"
	TEMPERATURE_LOW  = "temperature-low"
)

var Attributes = map[string]WeatherAttributes{
	TEMPERATURE: {Icon: "thermometer",
		Title:       "Temperature",
		Description: "temperature",
		Color:       "rgba(255,69,0, 255)",
		Chart:       "line",
		Selected:    true},
	FEELSLIKE: {Icon: "airwave",
		Title:       "Feels",
		Description: "feelslike",
		Color:       "rgba(255, 140, 0, 255)",
		Chart:       "line",
		Selected:    true},
	TEMPERATURE_HIGH: {Icon: "thermometer",
		Title:       "High",
		Description: "temperature high",
		Color:       "rgba(233,105,44, 255)",
		Chart:       "line",
		Selected:    true},
	TEMPERATURE_LOW: {Icon: "thermometer",
		Title:       "Low",
		Description: "temperature low",
		Color:       "rgba(255,179,71, 255)",
		Chart:       "line",
		Selected:    true},
	PRECIPITATION: {Icon: "weather_mix",
		Title:       "Precipitation",
		Description: "precipitation",
		Color:       "rgba(0,119,190)",
		Chart:       "bar",
		Selected:    true},
	PROBABILITY: {Icon: "weather_mix",
		Title:       "Probability",
		Description: "probability",
		Color:       "rgba(8,146,208, 255)",
		Chart:       "line",
		Selected:    true},
	RAIN: {Icon: "rainy",
		Title:       "Rain",
		Description: "rain",
		Color:       "rgba(31, 144, 255, 255)",
		Chart:       "line",
		Selected:    false},
	SHOWER: {Icon: "shower",
		Title:       "Showers",
		Description: "showers",
		Color:       "rgba(135,206,235, 255)",
		Chart:       "line",
		Selected:    false},
	SNOW: {Icon: "snowing",
		Title:       "Snow",
		Description: "snow",
		Color:       "rgba(255, 255, 255, 255)",
		Chart:       "line",
		Selected:    false},
	CLOUD: {Icon: "cloud",
		Title:       "Cloud",
		Description: "cloud cover",
		Color:       "rgba(119,139,165, 255)",
		Chart:       "line",
		Selected:    false},
	HUMIDITY: {Icon: "humidity_mid",
		Title:       "Humidity",
		Description: "humidity",
		Color:       "rgba(221,160,221, 255)",
		Chart:       "line",
		Selected:    false},
	WINDSPEED: {Icon: "toys_fan",
		Title:       "Wind Speed",
		Description: "wind speed",
		Color:       "rgba(255,218,185, 255)",
		Chart:       "line",
		Selected:    false},
	WINDGUSTS: {Icon: "air",
		Title:       "Gusts",
		Description: "wind gusts",
		Color:       "rgba(255,239,213, 255)",
		Chart:       "line",
		Selected:    false},
	PRESSURE: {Icon: "compress",
		Title:       "Pressure",
		Description: "pressure",
		Color:       "rgba(255,153,153, 255)",
		Chart:       "line",
		Selected:    false},
	SURFACE: {Icon: "compress",
		Title:       "Surface",
		Description: "surface pressure",
		Color:       "rgba(230,103,113, 255)",
		Chart:       "line",
		Selected:    false},
	DAYLIGHT: {Icon: "brightness_medium",
		Title:       "Daylight",
		Description: "daylight",
		Color:       "rgba(255,225,53, 255)",
		Chart:       "line",
		Selected:    false},
	SUNSHINE: {Icon: "brightness_7",
		Title:       "Sunshine",
		Description: "sunshine",
		Color:       "rgba(255, 255, 0, 255)",
		Chart:       "line",
		Selected:    false},
}
