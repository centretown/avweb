package runtime

import (
	"fmt"
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
	Webcams       map[string]*avcamx.AvItem
	Temp          *template.Template
}

// 45.41653080618134, -75.69649537375025
// https://api.open-meteo.com/v1/forecast?latitude=45.42&longitude=-75.70&hourly=temperature_2m
// "https://api.open-meteo.com/v1/forecast?latitude=45.42&longitude=-75.7&daily=sunrise,sunset&hourly=temperature_2m,apparent_temperature,precipitation_probability,precipitation,weather_code,wind_speed_10m&timezone=America%2FNew_York"

func NewRuntime(host *avcamx.AvHost) (rt *Runtime) {
	var webcamUrl = ""
	if len(host.Items) > 0 {
		webcamUrl = host.Items[0].Url
	}

	rt = &Runtime{
		Host:      host,
		WebcamUrl: webcamUrl,
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
		Webcams:   make(map[string]*avcamx.AvItem),
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
	for _, item := range rt.Host.Items {
		rt.Webcams[item.Url] = item
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
	Codes  any
}

func (rt *Runtime) HandleAction(path string, templ string, data *FormData) {

	rt.Host.Mux().HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if len(path) < 2 {
			return
		}

		w.Header().Add("Cache-Control", "no-cache")
		data.Action = rt.ActionMap[path[1:]]

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

	data := &FormData{
		Codes: WeatherCodes,
		Data:  rt.Locations}

	rt.HandleAction("/weather_sun", "weather.sun", data)
	rt.HandleAction("/weather_daily", "weather.daily", data)
	rt.HandleAction("/weather_hourly", "weather.hourly", data)

}

func (rt *Runtime) HandleCameras() {
	data := &FormData{
		Codes: avcamx.AvControllers["uvcvideo"],
		Data:  rt.Host}

	rt.HandleAction("/camera", "layout.controls", data)
	// rt.HandleAction("/cameraadd", "layout.camera.add", data)
	rt.HandleAction("/camera_list", "layout.camera.list", data)
	rt.Host.Mux().HandleFunc("/camera_primary", rt.setPrimaryCamera())
	rt.Host.Mux().HandleFunc("/record", rt.handleRecord())

}

func (rt *Runtime) setPrimaryCamera() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		const statusID = "camera_list_status"
		const sourceID = "source"

		wrapSource := func(id, src string) []byte {
			return []byte(fmt.Sprintf(`<img id="%s" src="%s">`, id, src))
		}

		cam, path, err := rt.parseCameraPath(r)
		if err != nil {
			msg := fmt.Sprintf("Error occured.<br>  %v", err)
			w.Write(wrapStatus(statusID, msg))
			return
		}

		if !cam.IsOpened() {
			msg := fmt.Sprintf("%s as %s is not connected", path, cam.Url)
			w.Write(wrapStatus(statusID, msg))
			return
		}

		msg := fmt.Sprintf("%s is connected as %s", path, cam.Url)
		w.Write(wrapStatus(statusID, msg))
		w.Write(wrapSource(sourceID, cam.Url))

		// `<img id="source" src="{{.WebcamUrl}}">`

	}
}

func wrapStatus(id, msg string) []byte {
	return []byte(fmt.Sprintf(`<div id="%s" class="status">%s</div>`, id, msg))
}

func (rt *Runtime) parseCameraPath(r *http.Request) (cam *avcamx.AvItem,
	path string, err error) {
	var (
		ok bool
	)
	err = r.ParseForm()
	if err != nil {
		err = fmt.Errorf("parse form: %v", err)
		return
	}

	path = r.FormValue("path")
	cam, ok = rt.Webcams[path]
	if !ok {
		err = fmt.Errorf("path not found: %s", path)
		return
	}
	return
}

func (rt *Runtime) handleRecord() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		avitem, err := rt.parseSourceId(r)
		if err != nil {
			return
		}

		if !avitem.IsRecording() {
			log.Printf("recording...")
			avitem.RecordCmd(300)
		} else {
			log.Printf("stop recording...")
			avitem.StopRecordCmd()
		}
	}
}

func (rt *Runtime) parseSourceId(r *http.Request) (item *avcamx.AvItem, err error) {
	err = r.ParseForm()
	if err != nil {
		log.Println("ParseForm", err)
		return
	}

	source := r.FormValue("source")
	item, ok := rt.Webcams[source]
	if !ok {
		err = fmt.Errorf("source: '%s' not found", source)
		return
	}
	return
}
