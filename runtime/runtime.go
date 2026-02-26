package runtime

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/centretown/avcamx"
	"github.com/centretown/avweb/action"
	"github.com/centretown/avweb/homeasst"
	"github.com/centretown/avweb/socket"
	"github.com/jmoiron/sqlx"
)

type Runtime struct {
	Location      *Location
	Locations     []*Location
	LocationIndex int
	WebcamUrl     string
	WebcamIndex   int
	ActionsCamera []*action.Action
	ActionsHome   []*action.Action
	ActionsChat   []*action.Action
	ActionMap     map[string]*action.Action
	WebSocket     *socket.Server
	Host          *avcamx.AvHost
	// Webcams   map[string]*avcamx.AvStream
	Template *template.Template
	Home     *homeasst.HomeRuntime
	ticker   *time.Ticker
	retry    *time.Ticker
	db       *sqlx.DB
}

func NewRuntime(host *avcamx.AvHost) (rt *Runtime) {
	var webcamUrl = ""
	streams := host.Streams()
	if len(streams) > 0 {
		webcamUrl = streams[0].Url
	}

	rt = &Runtime{
		Host:      host,
		WebcamUrl: webcamUrl,
		ActionsCamera: []*action.Action{
			{Name: "camera_list", Title: "Select Camera", Icon: "replace_video", Group: action.Camera},
			{Name: "camera", Title: "Camera Settings", Icon: "settings_video_camera", Group: action.Camera},
			// {Name: "cameraadd", Title: "Add Camera", Icon: "linked_camera", Group: Camera},
		},
		ActionsHome: []*action.Action{
			// {Name: "sun", Title: "Next Sun", Icon: "wb_twilight", Group: Home},
			{Name: "weather_current", Title: "Current Weather", Icon: "thunderstorm", Group: action.Home},
			{Name: "weather_hourly", Title: "24 Hour Forecast", Icon: "schedule", Group: action.Home},
			{Name: "weather_daily", Title: "7 Day Forecast", Icon: "calendar_view_week", Group: action.Home},
			// {Name: "weather_sun", Title: "Sun", Icon: "wb_twilight", Group: action.Home},
			{Name: "home", Title: "Home Assistant", Icon: "home", Group: action.Home},
			// {Name: "lights", Title: "LED Lights", Icon: "backlight_high", Group: action.Home},
		},
		ActionsChat: []*action.Action{
			// {Name: "chat", Title: "Chat", Icon: "chat", Group: Chat},
			{Name: "resetcontrols", Title: "Reset Camera", Icon: "reset_settings", Group: action.Chat},
			{Name: "record", Title: "Record", Icon: "radio_button_checked", Group: action.Chat},
		},

		ActionMap: make(map[string]*action.Action),
		// Webcams:   make(map[string]*avcamx.AvStream),
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

	rt.ticker = time.NewTicker(FirstTicker())

	err := rt.Connect()
	if err != nil {
		log.Fatal(err)
	}

	rt.Locations, err = SelectLocations(rt.db)
	if err != nil {
		log.Fatal(err)
	}
	rt.Location = rt.Locations[0]
	return
}

func (rt *Runtime) Connect() (err error) {
	rt.db, err = OpenDB("database/location.db")
	if err != nil {
		log.Print(err)
		return
	}
	return
}

func (rt *Runtime) SelectHistory(ID uint64, after string, before string) (history []*Current, err error) {
	return SelectHistoryInterval(rt.db, ID, after, before, "ASC")
}

func (rt *Runtime) LoadHistory() (err error) {
	after, before := BeforeTime(time.Now(), 6*time.Hour)
	for _, loc := range rt.Locations {
		history, err := rt.SelectHistory(loc.ID, after, before)
		if err != nil {
			log.Print(err)
		}
		loc.BuildCurrentProperties(history)
	}
	return
}

func (rt *Runtime) Done() {
	if rt.db != nil {
		rt.db.Close()
	}
}

type DailySummary struct {
	City            string
	High            string
	Low             string
	Precipitation   string
	Probability     string
	WindSpeed       string
	WindDirecection string
	WindGusts       string
	Code            string
	Color           string
}

func (rt *Runtime) CurrentWeatherDaily(index int) (hs DailySummary) {
	if index > len(rt.Locations) {
		return
	}

	loc := rt.Locations[index]
	daily := loc.WeatherDaily
	if len(daily.Daily.Time) < 1 {
		return
	}
	hs.City = loc.City
	hs.High = fmt.Sprintf("%4.1f %s",
		daily.Daily.High[0],
		daily.DailyUnits.High)
	hs.Low = fmt.Sprintf("%4.1f %s",
		daily.Daily.Low[0],
		daily.DailyUnits.Low)
	hs.Precipitation = fmt.Sprintf("%4.1f %s",
		daily.Daily.Precipitation[0],
		daily.DailyUnits.Precipitation)
	hs.Probability = fmt.Sprintf("%.0f%s",
		daily.Daily.Probability[0],
		daily.DailyUnits.Probability)
	code := WeatherCodes[daily.Daily.Code[0]]
	hs.Code = code.Icon
	hs.Color = code.Color
	return
}

type HourlySummary struct {
	City          string
	Temperature   string
	FeelsLike     string
	Precipitation string
	Probability   string
	WindSpeed     string
	WindDirection string
	WindGusts     string
	Humidity      string
	Pressure      string
	Code          string
	Color         string
}

func (rt *Runtime) CurrentTemperature() string {
	hourly := rt.Location.WeatherHourly
	if len(hourly.Hourly.Temperature) == 0 {
		return "99.9 ?"
	}
	return fmt.Sprintf("%2.1f%s",
		hourly.Hourly.Temperature[0],
		hourly.HourlyUnits.Temperature)
}

func (rt *Runtime) BroadcastTemperature() {
	buf := bytes.Buffer{}
	t := rt.Template.Lookup("weather.clock")
	t.Execute(&buf, rt)
	rt.WebSocket.Broadcast(buf.String())
}

type FormData struct {
	Action  *action.Action
	Data    any
	Codes   any
	Runtime *Runtime
}

type WeatherFormData struct {
	Action  *action.Action
	Data    any
	Codes   map[int32]*WeatherCode
	Runtime *Runtime
}

func (rt *Runtime) HandleAction(path string, templ string, data *WeatherFormData) {

	rt.Host.Mux().HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if len(path) < 2 {
			return
		}

		w.Header().Add("Cache-Control", "no-cache")
		data.Action = rt.ActionMap[path[1:]]

		err := rt.Template.Lookup(templ).Execute(w, data)
		if err != nil {
			log.Fatal(path, err)
		}
	})

}

func (rt *Runtime) HandleWeather() {
	data := &WeatherFormData{
		Codes:   WeatherCodes,
		Data:    rt.Locations,
		Runtime: rt}

	rt.HandleAction("/weather_daily", "weather.daily", data)
	rt.HandleAction("/weather_hourly", "weather.hourly", data)
	rt.HandleAction("/weather_current", "weather.current", data)

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

		err := rt.Template.Lookup(templ).Execute(w, data)
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
	// log.Printf("handleControl Replace req=%v", req)
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
	// log.Printf("handleControl Get req buf='%v'", string(buf))
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
			s := fmt.Sprintf(`<img id="%s" src="%s">`, id, src)
			return []byte(s)
		}

		cam, path, index, err := rt.parseCameraPath(r)
		if err != nil {
			msg := fmt.Sprintf("Error occured.<br>  %v", err)
			w.Write(wrapStatus(statusID, msg))
			return
		}

		rt.WebcamIndex = index
		rt.WebcamUrl = cam.Url

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
	var buf []byte
	buf = fmt.Appendf(buf, `<div id="%s" class="status">%s</div>`, id, msg)
	return buf
}

func (rt *Runtime) parseCameraPath(r *http.Request) (cam *avcamx.AvStream,
	path string, index int, err error) {
	err = r.ParseForm()
	if err != nil {
		err = fmt.Errorf("parse form: %v", err)
		return
	}

	path = r.FormValue("path")
	indexstr := r.FormValue("index")
	fmt.Sscanf(indexstr, "%d", &index)
	cam = rt.Host.Stream(path)
	if cam == nil {
		err = fmt.Errorf("path not found: %s", path)
		return
	}
	return
}

func (rt *Runtime) handleRecord() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		avStream, err := rt.parseSourceId(r)
		// _, err := rt.parseSourceId(r)
		if err != nil {
			log.Println("handleRecord", err)
			return
		}
		// log.Println("handleRecord", r.URL)

		if !avStream.IsRecording() {
			log.Printf("recording...")
			avStream.RecordCmd(3600)
		} else {
			log.Printf("stop recording...")
			avStream.StopRecordCmd()
		}
	}
}

func (rt *Runtime) parseSourceId(r *http.Request) (avStream *avcamx.AvStream, err error) {
	err = r.ParseForm()
	if err != nil {
		log.Println("ParseForm", err)
		return
	}

	source := r.FormValue("source")
	url := source[strings.LastIndex(source, "/"):]
	avStream = rt.Host.Stream(url)
	if avStream == nil {
		err = fmt.Errorf("url: '%s' not found", url)
		return
	}
	return
}

func (rt *Runtime) ServeHomeData() (err error) {
	home := rt.Home
	var ok bool
	ok, err = home.Authorize()
	if err != nil {
		log.Println("authorize", err)
		return
	}
	if !ok {
		err = fmt.Errorf("not authorized")
		log.Println(err)
		return
	}

	log.Println("Authorized HA")

	err = home.BuildEntities()
	if err != nil {
		log.Println("BuildEntities", err)
		return

	}
	log.Println("Build Entities")

	go home.Monitor()

	if home.Monitoring {
		log.Println("Monitor Entity States")
	}
	log.Println("Monitoring")

	return
}
