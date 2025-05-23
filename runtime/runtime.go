package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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
	Template      *template.Template

	ticker  *time.Ticker
	done    chan bool
	TickCmd chan int
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
			config.Actions["camera_list"],
			config.Actions["camera"],
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

		done:    make(chan bool),
		TickCmd: make(chan int),
	}

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
	ticker := next.Sub(now)
	log.Printf("ticker=%v, minute=%v now=%v next=%v\n", ticker, minute, now, next)
	rt.ticker = time.NewTicker(ticker)

	for _, item := range host.Items {
		rt.Webcams[item.Url] = item
	}

	rt.Locations = config.Locations
	rt.Location = rt.Locations[0]

	return
}

func (rt *Runtime) LoadHistory() error {
	buf, err := os.ReadFile("history.json")
	if err != nil {
		log.Println(err)
		return err
	}
	history := make([][]*Current, 0)
	err = json.Unmarshal(buf, &history)
	if err != nil {
		log.Println(err)
		return err
	}
	for i := range history {
		rt.Locations[i].History = history[i]
	}
	return err
}

func (rt *Runtime) SaveHistory() error {
	history := make([][]*Current, 0)
	for _, loc := range rt.Locations {
		history = append(history, loc.History)
	}
	buf, err := json.Marshal(history)
	if err != nil {
		log.Println(err)
		return err
	}
	err = os.WriteFile("history.json", buf, os.ModePerm)
	return err
}

func (rt *Runtime) Done() {
	// rt.WebSocket.Done()
	// rt.done <- true
}

func (rt *Runtime) Monitor() {
	const (
		NEXT_DAY = iota
		NEXT_HOUR
		NEXT_MINUTE
	)
	var (
		initialRun = true
	)

	log.Println("start Monitoring")
	defer func() {
		log.Println("stop Monitoring")
	}()

	for {
		select {
		case <-rt.done:
			log.Println("Done!")
			return

		case cmd := <-rt.TickCmd:
			switch cmd {
			case NEXT_DAY:
				rt.QueryDaily()
			case NEXT_HOUR:
				rt.QueryHourly()
				rt.BroadcastTemperature()
			case NEXT_MINUTE:
			}

		case now := <-rt.ticker.C:
			if initialRun {
				rt.ticker.Reset(time.Minute * 15)
				initialRun = false
			}
			log.Println("QueryCurrent")
			rt.QueryCurrent()
			rt.BroadcastTemperature()

			if now.Minute() == 0 {
				log.Println("QueryHourly")
				rt.QueryHourly()
				if now.Hour()%4 == 0 {
					log.Println("QueryDaily")
					rt.QueryDaily()
				}
			}
		}

		time.Sleep(time.Millisecond)
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
	Action  *Action
	Data    any
	Codes   any
	Runtime *Runtime
}

type WeatherFormData struct {
	Action  *Action
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

func (rt *Runtime) QueryDaily() {
	for i, location := range rt.Locations {
		err := location.QueryDaily()
		if err != nil {
			log.Printf("WeatherDaily: %v", err)
			continue
		}
		location.WeatherDaily.UpdateTime = time.Now()
		location.BuildDailyProperties(i)
	}
}

type LocationData struct {
	Index    int
	Location *Location
}

func (rt *Runtime) QueryHourly() {
	for i, location := range rt.Locations {
		err := location.QueryHourly()
		if err != nil {
			log.Printf("WeatherHourly: %v", err)
			continue
		}
		location.WeatherHourly.UpdateTime = time.Now()
		location.BuildHourlyProperties(i)
		// buf, _ := json.MarshalIndent(location.WeatherHourly, "", "  ")
		// log.Println(string(buf))
	}
}

func (rt *Runtime) QueryCurrent() {
	for i, location := range rt.Locations {
		err := location.QueryCurrent()
		if err != nil {
			log.Printf("WeatherCurrent: %v", err)
			continue
		}
		location.WeatherCurrent.UpdateTime = time.Now()
		location.BuildCurrentProperties(i)
	}
}

func (rt *Runtime) HandleWeather() {
	data := &WeatherFormData{
		Codes:   WeatherCodes,
		Data:    rt.Locations,
		Runtime: rt}

	rt.HandleAction("/weather_sun", "weather.sun", data)
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
