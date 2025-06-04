package runtime

import (
	"fmt"
	"log"
	"net/http"

	"github.com/centretown/avweb/action"
	"github.com/centretown/avweb/homeasst"
)

func (rt *Runtime) HandleHome() {
	mux := rt.Host.Mux()
	// mux.HandleFunc("/sun", rt.handleSun())
	// mux.HandleFunc("/weather", rt.handleWeather())
	mux.HandleFunc("/home", rt.HandleHomeAssistant())
	// mux.HandleFunc("/lights", rt.HandleLights())
	rt.HandleLightProperties()
}

type HomeAssistantData struct {
	Action   *action.Action
	Wifi     []*homeasst.WifiSensors
	Lights   []*homeasst.Light
	Entities homeasst.EntityMap
}

func (rt *Runtime) HandleHomeAssistant() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data := &HomeAssistantData{}
		data.Action = rt.ActionMap["home"]
		w.Header().Add("Cache-Control", "no-cache")
		data.Wifi = rt.Home.WifiSensors()
		data.Lights = rt.Home.NewLedLights()
		data.Entities = rt.Home.Entities
		err := rt.Template.Lookup("layout.home").Execute(w, data)
		if err != nil {
			log.Fatal("/home", err)
		}
	}
}

func (rt *Runtime) HandleLights() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache")
		lights := rt.Home.NewLedLights()
		err := rt.Template.Lookup("layout.lights").Execute(w, lights)
		if err != nil {
			log.Fatal("/lights", err)
		}
	}
}

func (rt *Runtime) HandleLightProperties() {
	home := rt.Home
	mux := rt.Host.Mux()
	mux.HandleFunc("/light/state",
		func(w http.ResponseWriter, r *http.Request) {
			log.Println("/light/state")
			id, key, _ := ReadBody(r)
			if key == "state" {
				home.CallService(LightCmd(id))
			} else {
				home.CallService(LightCmdOff(id))
			}
		})

	mux.HandleFunc("/light/brightness",
		func(w http.ResponseWriter, r *http.Request) {
			log.Println("/light/brightness")
			id, key, val := ReadBody(r)
			home.CallService(LightCmd(id, ServiceData{Key: key, Value: val}))
		})

	mux.HandleFunc("/light/color",
		func(w http.ResponseWriter, r *http.Request) {
			log.Println("/light/color")
			id, key, val := ReadBody(r)
			length := len(val)
			if length > 6 {
				val := val[length-6:]
				var red, green, blue int
				fmt.Sscan(fmt.Sprintf("0x%s 0x%s 0x%s", val[:2], val[2:4], val[4:]),
					&red, &green, &blue)
				home.CallService(LightCmd(id, ServiceData{Key: key,
					Value: fmt.Sprintf("[%d,%d,%d]", red, green, blue)}))
			}
		})

	mux.HandleFunc("/light/effect",
		func(w http.ResponseWriter, r *http.Request) {
			log.Println("/light/effect")
			id, key, val := ReadBody(r)
			home.CallService(LightCmd(id, ServiceData{Key: key,
				Value: `"` + val + `"`}))
		})
}
