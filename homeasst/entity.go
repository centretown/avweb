package homeasst

import (
	"encoding/json"
	"time"
)

type Entity[T any] struct {
	EntityID     string    `json:"entity_id" yaml:"entity_id"`
	State        string    `json:"state" yaml:"state"`
	LastChanged  time.Time `json:"last_changed" yaml:"last_changed"`
	LastReported time.Time `json:"last_reported" yaml:"last_reported"`
	LastUpdated  time.Time `json:"last_updated" yaml:"last_updated"`
	Context      Context   `json:"context" yaml:"context"`
	Attributes   T         `json:"attributes" yaml:"attributes"`
}

type EntityMap map[string]*Entity[json.RawMessage]

type NumberAttributes struct {
	Min   float64 `json:"min" yaml:"min"`
	Max   float64 `json:"max" yaml:"max"`
	Step  float64 `json:"step" yaml:"step"`
	Mode  string  `json:"mode" yaml:"mode"`
	Units string  `json:"unit_of_measurement" yaml:"unit_of_measurement"`
	Name  string  `json:"friendly_name" yaml:"friendly_name"`
}

// weather.forecast_home
type WeatherAttributes struct {
	Attribution       string  `json:"attribution" yaml:"attribution"`
	CloudCoverage     float64 `json:"cloud_coverage" yaml:"cloud_coverage"`
	DewPoint          float64 `json:"dew_point" yaml:"dew_point"`
	Name              string  `json:"friendly_name" yaml:"friendly_name"`
	Humidity          float64 `json:"humidity" yaml:"humidity"`
	PrecipitationUnit string  `json:"precipitation_unit" yaml:"precipitation_unit"`
	Pressure          float64 `json:"pressure" yaml:"pressure"`
	PressureUnit      string  `json:"pressure_unit" yaml:"pressure_unit"`
	SupportedFeatures int     `json:"supported_features" yaml:"supported_features"`
	Temperature       float64 `json:"temperature" yaml:"temperature"`
	TemperatureUnit   string  `json:"temperature_unit" yaml:"temperature_unit"`
	VisibilityUnit    string  `json:"visibility_unit" yaml:"visibility_unit"`
	WindBearing       float64 `json:"wind_bearing" yaml:"wind_bearing"`
	WindSpeed         float64 `json:"wind_speed" yaml:"wind_speed"`
	WindSpeedUnit     string  `json:"wind_speed_unit" yaml:"wind_speed_unit"`
}

// sensor.wifi_signal_28

type SensorAttributes struct {
	StateClass  string `json:"state_class" yaml:"state_class"`
	Units       string `json:"unit_of_measurement" yaml:"unit_of_measurement"`
	DeviceClass string `json:"device_class" yaml:"device_class"`
	Name        string `json:"friendly_name" yaml:"friendly_name"`
}

type ZoneAttributes struct {
	Latitude  float64  `json:"latitude" yaml:"latitude"`
	Longitude float64  `json:"longitude" yaml:"longitude"`
	Radius    float64  `json:"radius" yaml:"radius"`
	Passive   bool     `json:"passive" yaml:"passive"`
	Persons   []string `json:"persons" yaml:"persons"`
	Editable  bool     `json:"editable" yaml:"editable"`
	Icon      string   `json:"icon" yaml:"icon"`
	Name      string   `json:"friendly_name" yaml:"friendly_name"`
}

type TimeStampAttributes struct {
	DeviceClass string `json:"device_class" yaml:"device_class"`
	Name        string `json:"friendly_name" yaml:"friendly_name"`
}

type Number struct {
	Entity[NumberAttributes]
}
type AnyData struct {
	Entity[any]
}
type Zone struct {
	Entity[ZoneAttributes]
}

func (ent *Entity[T]) Copy(src *Entity[json.RawMessage]) {
	ent.EntityID = src.EntityID
	ent.State = src.State
	ent.LastChanged = src.LastChanged
	ent.LastUpdated = src.LastUpdated
	ent.LastReported = src.LastReported
	ent.Context = src.Context
	json.Unmarshal(src.Attributes, &ent.Attributes)
}

func (ent *Entity[T]) FormatTime() string {
	return ent.LastUpdated.Local().Format(ShortTime)
}
