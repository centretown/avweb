package runtime

import (
	"html/template"
	"log"
	"net/http"

	"github.com/centretown/avcamx"
	"github.com/centretown/avweb/socket"
)

type Runtime struct {
	Location      *Location
	Locations     []*Location
	WebcamUrl     string
	ActionsCamera []*Action
	ActionsHome   []*Action
	ActionsChat   []*Action
	ActionMap     map[string]*Action
	WebSocket     *socket.Server
	Host          *avcamx.AvHost
	Temp          *template.Template
}

// 45.41653080618134, -75.69649537375025
// https://api.open-meteo.com/v1/forecast?latitude=45.42&longitude=-75.70&hourly=temperature_2m
// "https://api.open-meteo.com/v1/forecast?latitude=45.42&longitude=-75.7&daily=sunrise,sunset&hourly=temperature_2m,apparent_temperature,precipitation_probability,precipitation,weather_code,wind_speed_10m&timezone=America%2FNew_York"

func NewRuntime() (rt *Runtime) {
	rt = &Runtime{
		ActionsCamera: []*Action{
			{Name: "camera", Title: "Camera Settings", Icon: "settings_video_camera", Group: Camera},
			{Name: "cameraadd", Title: "Add Camera", Icon: "linked_camera", Group: Camera},
			{Name: "camera_list", Title: "List Cameras", Icon: "view_list", Group: Camera},
		},
		ActionsHome: []*Action{
			// {Name: "sun", Title: "Next Sun", Icon: "wb_twilight", Group: Home},
			{Name: "weather_hourly", Title: "Hourly Forecast", Icon: "thermometer", Group: Home},
			{Name: "weather_daily", Title: "Daily Forecast", Icon: "calendar_view_week", Group: Home},
			{Name: "weather_sun", Title: "Sun", Icon: "wb_twilight", Group: Home},
			// {Name: "wifi", Title: "WIFI Signals", Icon: "network_wifi", Group: Home},
			// {Name: "lights", Title: "LED Lights", Icon: "backlight_high", Group: Home},
		},
		ActionsChat: []*Action{
			// {Name: "chat", Title: "Chat", Icon: "chat", Group: Chat},
			{Name: "resetcontrols", Title: "Reset Camera", Icon: "reset_settings", Group: Chat},
			{Name: "record", Title: "Record", Icon: "radio_button_checked", Group: Chat},
		},

		ActionMap: make(map[string]*Action),
	}

	for _, action := range rt.ActionsCamera {
		rt.ActionMap[action.Name] = action
	}
	for _, action := range rt.ActionsHome {
		rt.ActionMap[action.Name] = action
	}
	for _, action := range rt.ActionsChat {
		rt.ActionMap[action.Name] = action
	}

	//45.40608984676536, -75.68631292544273
	//48.485340413458964, -81.33687676821839
	//49.77255342394314, -94.48309874840045
	rt.Locations = []*Location{
		{City: "Ottawa", Latitude: 45.40608984676536, Longitude: -75.68631292544273, Zone: "America%2FNew_York"},
		{City: "Timmins", Latitude: 48.485340413458964, Longitude: -81.33687676821839, Zone: "America%2FNew_York"},
		{City: "Kenora", Latitude: 49.77255342394314, Longitude: -94.48309874840045, Zone: "America%2FNew_York"},
	}
	rt.Location = rt.Locations[0]
	return
}

type FormData struct {
	Action *Action
	Data   any
}

func (rt *Runtime) HandleAction(path string, templ string, data any) {

	rt.Host.Mux().HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if len(path) < 2 {
			return
		}

		w.Header().Add("Cache-Control", "no-cache")
		data := &FormData{
			Action: rt.ActionMap[path[1:]],
			Data:   data}

		err := rt.Temp.Lookup(templ).Execute(w, data)
		if err != nil {
			log.Fatal(path, err)
		}
	})

}

func (rt *Runtime) HandleWeather() {

	for _, location := range rt.Locations {
		err := location.QueryDaily()
		if err != nil {
			log.Printf("WeatherDaily: %v", err)
		}
		err = location.QueryHourly()
		if err != nil {
			log.Printf("WeatherHourly: %v", err)
		}
	}

	rt.HandleAction("/weather_sun", "weather.sun", rt.Locations)
	rt.HandleAction("/weather_daily", "weather.daily", rt.Locations)
	rt.HandleAction("/weather_hourly", "weather.hourly", rt.Locations)

}
