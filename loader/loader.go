package main

import (
	"log"

	"github.com/centretown/avweb/runtime"
)

var ActionsCamera = []*runtime.Action{
	{Name: "camera", Title: "Camera Settings", Icon: "settings_video_camera", Group: runtime.Camera},
	// {Name: "cameraadd", Title: "Add Camera", Icon: "linked_camera", Group: runtime.Camera},
	{Name: "camera_list", Title: "List Cameras", Icon: "view_list", Group: runtime.Camera},
}
var ActionsHome = []*runtime.Action{
	// {Name: "sun", Title: "Next Sun", Icon: "wb_twilight", Group: Home},
	{Name: "weather_current", Title: "Current Readings", Icon: "thunderstorm", Group: runtime.Home},
	{Name: "weather_hourly", Title: "24 Hour Forecast", Icon: "schedule", Group: runtime.Home},
	{Name: "weather_daily", Title: "7 Day Forecast", Icon: "calendar_view_week", Group: runtime.Home},
	{Name: "weather_sun", Title: "Sun", Icon: "wb_twilight", Group: runtime.Home},
	// {Name: "wifi", Title: "WIFI Signals", Icon: "network_wifi", Group: Home},
	// {Name: "lights", Title: "LED Lights", Icon: "backlight_high", Group: Home},
}

var ActionsChat = []*runtime.Action{
	// {Name: "chat", Title: "Chat", Icon: "chat", Group: Chat},
	{Name: "resetcontrols", Title: "Reset Camera", Icon: "reset_settings", Group: runtime.Chat},
	{Name: "record", Title: "Record", Icon: "radio_button_checked", Group: runtime.Chat},
}

var Locations = []*runtime.Location{
	{City: "Ottawa", Latitude: 45.40608984676536, Longitude: -75.68631292544273, Zone: "America%2FNew_York"},
	{City: "Timmins", Latitude: 48.485340413458964, Longitude: -81.33687676821839, Zone: "America%2FNew_York"},
	{City: "Kenora", Latitude: 49.77255342394314, Longitude: -94.48309874840045, Zone: "America%2FNew_York"},
	{City: "Edmonton", Latitude: 53.529490711683785, Longitude: -113.50925706225637, Zone: "America%2FNew_York"},
}

func main() {
	filename := "config.json"
	cfg := runtime.NewConfig()
	for _, action := range ActionsCamera {
		cfg.Actions[action.Name] = action
	}
	for _, action := range ActionsHome {
		cfg.Actions[action.Name] = action
	}
	for _, action := range ActionsChat {
		cfg.Actions[action.Name] = action
	}
	cfg.Locations = Locations

	err := cfg.Write(filename)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("%s was successfully created\n", filename)
}
