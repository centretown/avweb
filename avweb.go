package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"

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

	log.Print("NEW AVHOST")
	host = avcamx.NewAvHost(hostAddr, hostPort)
	host.MakeLocal()

	log.Print("FetchRemote")
	remote, err := host.FetchRemote(remoteAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("MakeProxy")
	host.MakeProxy(remote)

	rt = runtime.NewRuntime()
	rt.WebcamUrl = host.Items[0].Url
	rt.Host = host
	mux := host.Mux()
	const pattern = "www/*.html"
	rt.Temp, err = template.ParseGlob(pattern)
	if err != nil {
		log.Fatalln("ParseGlob", pattern, err)
	}

	rt.WebSocket = socket.NewServer(rt.Temp)
	rt.WebSocket.LoadMessages()
	rt.WebSocket.Run()

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

		rt.Temp.ExecuteTemplate(w, "index.html", rt)
	})

	mux.HandleFunc("/events", rt.WebSocket.Events)

	rt.HandleWeather()

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

	host.Quit()
}
