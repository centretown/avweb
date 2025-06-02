package homeasst

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/centretown/avweb/sockclient"
)

const (
	AuthCommand      = `{ "type":"auth", "access_token":"%s" }`
	ConfigCommand    = `{ "type":"get_config", "id":%d }`
	StatesCommand    = `{ "type":"get_states", "id":%d }`
	SubscribeCommand = `{ "type":"subscribe_events", "event_type":"state_changed", "id":%d }`
)

type HomeRuntime struct {
	Entities   EntityMap
	EntityKeys []string

	Monitoring      bool
	Temperature     float64
	TemperatureUnit string

	stop          chan int
	loadStatesID  int
	subscriptions map[string][]*Subscription
	eventsID      int
	sock          *sockclient.SockClient
}

func NewHomeRuntime() (*HomeRuntime, error) {
	var home = &HomeRuntime{
		Entities:      make(map[string]*Entity[json.RawMessage]),
		stop:          make(chan int),
		subscriptions: make(map[string][]*Subscription),
	}

	var err error
	home.sock, err = sockclient.NewSockClient()
	if err != nil {
		log.Println("NewHomeData", err)
	}
	return home, err
}

func (home *HomeRuntime) Subscribe(entityID string, subscription *Subscription) {
	list, ok := home.subscriptions[entityID]
	if !ok {
		list = make([]*Subscription, 1)
		list[0] = subscription
	} else {
		list = append(list, subscription)
	}
	home.subscriptions[entityID] = list
}

func (home *HomeRuntime) Consume(entityID string, newState *Entity[json.RawMessage]) {
	subs, ok := home.subscriptions[entityID]
	if ok {
		for _, sub := range subs {
			if sub.Enabled {
				sub.Consume(newState)
			}
		}
	}
}

func (home *HomeRuntime) EntityList(filters ...string) (list []string) {
	list = make([]string, 0)
	all := len(filters) == 0
	for keys := range home.Entities {
		if all {
			list = append(list, keys)
			continue
		}
		for _, filter := range filters {
			if strings.HasPrefix(keys, filter) {
				list = append(list, keys)
				break
			}
		}
	}
	slices.Sort(list)
	return
}

func (home *HomeRuntime) CallService(cmd string) {
	home.sock.WriteCommandID(cmd)
}

func (home *HomeRuntime) StopMonitor() {
	if home.Monitoring {
		log.Println("StopMonitor")
		home.stop <- 1
	}
}

func (home *HomeRuntime) Authorize() (ok bool, err error) {
	var (
		result AuthResult
		buf    []byte
		max    = 4
	)

	cmd := fmt.Sprintf(AuthCommand, sockclient.Token)

	for i := 0; i < max && !ok; i += 1 {
		err = home.sock.WriteCommand(cmd)
		if err != nil {
			return
		}
		buf, err = home.sock.Read()
		if err != nil {
			return
		}

		err = json.Unmarshal(buf, &result)
		if err != nil {
			return
		}

		ok = result.Type == "auth_ok"
	}

	return
}

func (home *HomeRuntime) BuildEntities() (err error) {
	home.loadStatesID, err = home.sock.WriteCommandID(StatesCommand)
	if err != nil {
		log.Println("BuildEntities", err)
		return
	}
	var (
		buf []byte
	)

	log.Println("BuildEntities Loop")
	for {
		buf, err = home.sock.Read()
		if err != nil && err != io.EOF {
			log.Println("ReadEntities", err)
			return
		}

		if len(buf) > 0 {
			home.ParseResponse(buf)
			if len(home.Entities) > 0 {
				log.Println("COUNT", len(buf), len(home.Entities))
				break
			}
			log.Println(len(buf), string(buf))
		}
		time.Sleep(time.Millisecond)
	}

	home.EntityKeys = BuildEntityKeys(home.Entities)
	return
}

func (home *HomeRuntime) Monitor() {
	log.Println("monitor")
	var (
		errCount int
		delay    time.Duration = time.Millisecond * 5
		err      error
	)

	home.Monitoring = false
	home.loadStatesID, err = home.sock.WriteCommandID(StatesCommand)
	if err != nil {
		log.Println("StatesCommand", err)
		return
	}
	log.Println("StatesCommand")
	home.Monitoring = true

	home.eventsID, err = home.sock.WriteCommandID(SubscribeCommand)
	if err != nil {
		log.Println("SubscribeCommand", err)
		return
	}
	log.Println("SubscribeCommand")

	for {
		time.Sleep(delay)

		select {
		case <-home.stop:
			log.Println("STOP RECEIVED")
			home.Monitoring = false
			return

		default:
			buf, err := home.sock.Read()
			if err != nil {
				errCount++
				if errCount > 10 {
					log.Fatal(err)
				}
				log.Println(err)
				continue
			}
			errCount = 0
			if len(buf) > 0 {
				go home.ParseResponse(buf)
			}
		}
	}
}
