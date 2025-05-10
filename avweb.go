package main

import (
	"context"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/centretown/avcamx"
	"github.com/centretown/avweb/runtime"
	"github.com/centretown/avweb/socket"
)

func main() {
	var (
		remoteAddr      = "http://192.168.10.197:8080"
		remoteAddrUsage = "remote camera ip address"
		hostAddr        = avcamx.GetOutboundIP()
		hostAddrUsage   = "web site ip address"
		hostPort        = "9000"
		hostPortUsage   = "web site ip port number"
		host            *avcamx.AvHost
		rt              *runtime.Runtime
	)

	flag.StringVar(&hostAddr, "host", hostAddr, hostAddrUsage)
	flag.StringVar(&hostAddr, "h", hostAddr, hostAddrUsage)
	flag.StringVar(&hostPort, "port", hostPort, hostPortUsage)
	flag.StringVar(&hostPort, "p", hostPort, hostPortUsage)
	flag.StringVar(&remoteAddr, "remote", remoteAddr, remoteAddrUsage)
	flag.StringVar(&remoteAddr, "r", remoteAddr, remoteAddrUsage)
	flag.Parse()

	filename := "config.json"
	config := runtime.NewConfig()
	err := config.Read(filename)
	if err != nil {
		log.Fatalln("Read Configuration", filename, err)
	}

	host = avcamx.NewAvHost(hostAddr, hostPort)
	const pattern = "www/*.html"
	templ, err := template.ParseGlob(pattern)
	if err != nil {
		log.Fatalln("ParseGlob", pattern, err)
	}
	sockServer := socket.NewServer(templ)

	host.MakeLocal(sockServer)
	log.Print("FetchRemote")
	remote, err := host.FetchRemote(remoteAddr)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("MakeProxy")
	host.MakeProxy(remote, sockServer)

	rt = runtime.NewRuntime(host, config)
	rt.Temp = templ
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
		w.Header().Add("Cache-Control", "no-cache")
		rt.Temp, err = template.ParseGlob(pattern)
		if err != nil {
			log.Fatalln("ParseGlob", pattern, err)
		}
		rt.WebSocket.UpdateTemplate(rt.Temp)
		rt.Temp.ExecuteTemplate(w, "index.html", rt)
	})

	mux.HandleFunc("/events", rt.WebSocket.Events)
	mux.HandleFunc("/msghook", rt.WebSocket.MessageHook)

	rt.QueryDaily()
	rt.QueryHourly()

	rt.HandleWeather()
	rt.HandleCameras()

	go rt.Monitor()

	httpErr := make(chan error, 1)
	go func() {
		httpErr <- host.ListenAndServe()
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case err := <-httpErr:
		log.Printf("failed to serve http: %v", err)
	case sig := <-sigs:
		log.Printf("terminating: %v", sig)
	}

	rt.WebSocket.SaveMessages()

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Second)
	defer cancel()

	host.Quit()
	host.Server.Shutdown(ctx)

	rt.Done()
}
