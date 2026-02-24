package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/centretown/avcamx"
	"github.com/centretown/avweb/homeasst"
	"github.com/centretown/avweb/runtime"
	"github.com/centretown/avweb/socket"
)

var funcMap = template.FuncMap{
	"sub": func(i int, j int) int {
		return i - j
	},
}

func main() {
	avFlags := avcamx.NewAvFlags()
	exists := avFlags.HasFile()
	if exists {
		avFlags.Load()
	}

	avFlags.Parse()
	err := avFlags.Save()
	if err != nil {
		log.Printf("Error saving configuration file %s. %s", avcamx.ConfigName, err)
	} else if exists {
		log.Print("Saved configuration file. ", avcamx.ConfigName)
	} else {
		log.Print("Created configuration file. ", avcamx.ConfigName)
	}

	avFlags.Print()

	const pattern = "www/*.html"
	templ, err := template.New("").Funcs(funcMap).ParseGlob(pattern)
	if err != nil {
		log.Fatalln("ParseGlob", pattern, err)
	}

	sockServer := socket.NewServer(templ)
	var listener avcamx.StreamListener = sockServer
	host := avcamx.NewAvHost(avFlags.HostAddr, avFlags.HostPort, avFlags.Remotes, 1000, listener)
	// time.Sleep(time.Second * 2)
	log.Printf("\nServing %s...", host.Url)

	rt := runtime.NewRuntime(host)

	rt.Template = templ
	rt.WebSocket = sockServer
	rt.WebSocket.LoadMessages()
	rt.WebSocket.Run()

	mux := host.Mux()

	fs := http.FileServer(http.Dir("www/"))
	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache")
		http.StripPrefix("/static/", fs).ServeHTTP(w, r)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// refresh template (dev only)
		w.Header().Add("Cache-Control", "no-cache")
		rt.Template, err = template.New("").Funcs(funcMap).ParseGlob(pattern)
		if err != nil {
			log.Fatalln("ParseGlob", pattern, err)
		}
		rt.WebSocket.UpdateTemplate(rt.Template)
		rt.Template.ExecuteTemplate(w, "index.html", rt)
	})

	mux.HandleFunc("/events", rt.WebSocket.Events)
	mux.HandleFunc("/msghook", rt.WebSocket.MessageHook)

	rt.QueryCurrent()
	rt.QueryDaily()
	rt.QueryHourly()

	rt.HandleWeather()
	rt.HandleCameras()

	rt.Home, err = homeasst.NewHomeRuntime()
	if err == nil {
		rt.ServeHomeData()
		rt.HandleHome()
	}

	go rt.Monitor()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	sig := <-sigs
	log.Printf("Signal: %v", sig)

	rt.WebSocket.SaveMessages()
	// rt.SaveHistory()

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Second)
	defer cancel()

	host.Quit()
	host.Server.Shutdown(ctx)

	rt.Done()
}
