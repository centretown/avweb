package runtime

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/centretown/avcamx"
	"github.com/centretown/avweb/socket"
)

type Runtime struct {
	Location      *Location
	Locations     []*Location
	LocationIndex int
	WebcamUrl     string
	WebcamIndex   int
	ActionsCamera []*Action
	ActionsHome   []*Action
	ActionsChat   []*Action
	ActionMap     map[string]*Action
	WebSocket     *socket.Server
	Host          *avcamx.AvHost
	Webcams       map[string]*avcamx.AvItem
	Temp          *template.Template
}

func NewRuntime(host *avcamx.AvHost, config *Config) (rt *Runtime) {
	var webcamUrl = ""
	if len(host.Items) > 0 {
		webcamUrl = host.Items[0].Url
	}

	rt = &Runtime{
		Host:      host,
		WebcamUrl: webcamUrl,
		ActionsCamera: []*Action{
			config.Actions["camera"],
			config.Actions["camera_list"],
		},
		ActionsHome: []*Action{
			config.Actions["weather_current"],
			config.Actions["weather_hourly"],
			config.Actions["weather_daily"],
			config.Actions["weather_sun"],
		},
		ActionsChat: []*Action{
			config.Actions["resetcontrols"],
			config.Actions["record"],
		},

		ActionMap: config.Actions,
		Webcams:   make(map[string]*avcamx.AvItem),
	}

	for _, item := range host.Items {
		rt.Webcams[item.Url] = item
	}

	rt.Locations = config.Locations
	rt.Location = rt.Locations[0]
	return
}

type FormData struct {
	Action  *Action
	Data    any
	Codes   any
	Runtime *Runtime
}

type WeatherFormData struct {
	Action  *Action
	Data    any
	Codes   map[int]*WeatherCode
	Runtime *Runtime
}

func (rt *Runtime) HandleAction(path string, templ string, data *WeatherFormData) {

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

	data := &WeatherFormData{
		Codes:   WeatherCodes,
		Data:    rt.Locations,
		Runtime: rt}

	rt.HandleAction("/weather_current", "weather.current", data)
	rt.HandleAction("/weather_sun", "weather.sun", data)
	rt.HandleAction("/weather_daily", "weather.daily", data)
	rt.HandleAction("/weather_hourly", "weather.hourly", data)

}

func (rt *Runtime) HandleCameraAction(path string, templ string) {

	rt.Host.Mux().HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if len(path) < 2 {
			return
		}

		w.Header().Add("Cache-Control", "no-cache")
		data := &FormData{
			Codes:   avcamx.AvControllers["uvcvideo"],
			Runtime: rt,
			Data:    rt.Host}

		data.Action = rt.ActionMap[path[1:]]

		err := rt.Temp.Lookup(templ).Execute(w, data)
		if err != nil {
			log.Fatal(path, err)
		}
	})

}

func (rt *Runtime) handleControl(w http.ResponseWriter, r *http.Request) {
	url := r.URL.String()
	r.ParseForm()
	defer r.Body.Close()
	source := r.FormValue("source")
	req := strings.Replace(url, "/camera_control", source, 1)
	resp, err := http.Get(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	w.Write(buf)
}

func (rt *Runtime) HandleCameras() {
	rt.HandleCameraAction("/camera", "layout.controls")
	rt.HandleCameraAction("/camera_list", "layout.camera.list")
	rt.Host.Mux().HandleFunc("/camera_primary", rt.setPrimaryCamera())
	rt.Host.Mux().HandleFunc("/record", rt.handleRecord())
	rt.Host.Mux().HandleFunc("/camera_control/", rt.handleControl)

}

func (rt *Runtime) setPrimaryCamera() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		const statusID = "camera_list_status"
		const sourceID = "source"

		wrapSource := func(id, src string) []byte {
			return []byte(fmt.Sprintf(`<img id="%s" src="%s">`, id, src))
		}

		cam, path, index, err := rt.parseCameraPath(r)
		if err != nil {
			msg := fmt.Sprintf("Error occured.<br>  %v", err)
			w.Write(wrapStatus(statusID, msg))
			return
		}

		rt.WebcamIndex = index
		rt.WebcamUrl = rt.Host.Items[index].Url

		if !cam.IsOpened() {
			msg := fmt.Sprintf("%s as %s is not connected", path, cam.Url)
			w.Write(wrapStatus(statusID, msg))
			return
		}

		msg := fmt.Sprintf("%s is connected as %s (%d)", path, cam.Url, index)
		w.Write(wrapStatus(statusID, msg))
		w.Write(wrapSource(sourceID, cam.Url))

		// `<img id="source" src="{{.WebcamUrl}}">`

	}
}

func wrapStatus(id, msg string) []byte {
	return []byte(fmt.Sprintf(`<div id="%s" class="status">%s</div>`, id, msg))
}

func (rt *Runtime) parseCameraPath(r *http.Request) (cam *avcamx.AvItem,
	path string, index int, err error) {
	var (
		ok bool
	)
	err = r.ParseForm()
	if err != nil {
		err = fmt.Errorf("parse form: %v", err)
		return
	}

	path = r.FormValue("path")
	indexstr := r.FormValue("index")
	fmt.Sscanf(indexstr, "%d", &index)
	cam, ok = rt.Webcams[path]
	if !ok {
		err = fmt.Errorf("path not found: %s", path)
		return
	}
	log.Println("parseCameraPath", path, indexstr, index)
	return
}

func (rt *Runtime) handleRecord() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// avitem, err := rt.parseSourceId(r)
		_, err := rt.parseSourceId(r)
		if err != nil {
			log.Println("handleRecord", err)
			return
		}
		log.Println("handleRecord", r.URL)

		// 	if !avitem.IsRecording() {
		// 		log.Printf("recording...")
		// 		avitem.RecordCmd(300)
		// 	} else {
		// 		log.Printf("stop recording...")
		// 		avitem.StopRecordCmd()
		// 	}
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
