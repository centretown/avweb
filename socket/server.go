package socket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/centretown/avcamx"
)

type Server struct {
	messageChan     chan string
	Messages        []*Message
	mutex           sync.Mutex
	hub             *Hub
	statusLayout    *template.Template
	messageLayout   *template.Template
	activeRecorders int
}

func NewServer(t *template.Template) *Server {
	s := &Server{
		messageChan:   make(chan string),
		Messages:      make([]*Message, 0),
		hub:           NewHub(),
		statusLayout:  t.Lookup("layout.wsstatus"),
		messageLayout: t.Lookup("layout.wsmessage"),
	}
	return s
}

func (s *Server) UpdateTemplate(t *template.Template) {
	s.statusLayout = t.Lookup("layout.wsstatus")
	s.messageLayout = t.Lookup("layout.wsmessage")
}

func (s *Server) Run() {
	go s.hub.Run()
}

func (s *Server) PastMessages() (past []*Message) {
	max := len(s.Messages)
	past = make([]*Message, max)
	for i := range max {
		past[max-i-1] = s.Messages[i]
	}
	return
}

const messageFile = "messages.json"

func (s *Server) LoadMessages() (err error) {
	var buf []byte
	buf, err = os.ReadFile(messageFile)
	if err != nil {
		log.Println("LoadMessages:ReadFile", err)
		return
	}

	err = json.Unmarshal(buf, &s.Messages)
	if err != nil {
		log.Println("LoadMessages:Unmarshal", err)
		return
	}
	return
}

func (s *Server) SaveMessages() (err error) {
	var buf []byte
	buf, err = json.MarshalIndent(s.Messages, "", "  ")
	if err != nil {
		log.Println("SaveMessages:MarshalIndent", err)
		return
	}
	err = os.WriteFile(messageFile, buf, os.ModePerm)
	if err != nil {
		log.Println("SaveMessages:WriteFile", err)
		return
	}
	return
}

var _ avcamx.StreamListener = (*Server)(nil)

const (
	streamOff     = `<span id="streamer" hx-swap-oob="outerHTML" class="symbols">radio_button_checked</span>`
	streamOn      = `<span id="streamer" hx-swap-oob="outerHTML" class="symbols streaming">radio_button_checked</span>`
	streamOnList  = `<span id="stream_video%d" hx-swap-oob="outerHTML" class="symbols-form streaming">radio_button_checked</span>`
	streamOffList = `<span id="stream_video%d" hx-swap-oob="outerHTML" class="symbols-form">radio_button_checked</span>`
)

func (s *Server) StreamOn(id int) {
	s.activeRecorders += 1
	s.Broadcast(streamOn)
	s.Broadcast(fmt.Sprintf(streamOnList, id))
}

func (s *Server) StreamOff(id int) {
	s.activeRecorders -= 1
	if s.activeRecorders == 0 {
		s.Broadcast(streamOff)
	}
	s.Broadcast(fmt.Sprintf(streamOffList, id))
}

func (s *Server) Broadcast(message string) {
	s.hub.Broadcast(message)
}

func (s *Server) MessageHook(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Printf("Failed to parse body: %v", err)
		return
	}

	name := "unknown"
	names, ok := r.PostForm["name"]
	if ok && len(names) > 0 {
		name = names[0]
	}

	message := "empty"
	messages, ok := r.PostForm["message"]
	if ok && len(messages) > 0 {
		message = messages[0]
	}

	log.Printf("Received webhook: %s %s", name, message)

	var (
		buf bytes.Buffer
		msg = &Message{Name: name, Message: message, Stamp: time.Now()}
	)

	err = s.messageLayout.Execute(&buf, msg)
	if err != nil {
		log.Printf("Failed to execute template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// wrap the message in a div so we can use htmx to add it to the page
	s.hub.broadcast <- []byte("<div hx-swap-oob=\"afterbegin:#messages\">" + buf.String() + "</div>")

	s.mutex.Lock()
	s.Messages = append(s.Messages, msg)
	if len(s.Messages) > maxMessages {
		s.Messages = s.Messages[1:]
	}
	log.Printf("Now have %d past messages", len(s.Messages))
	s.mutex.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (s *Server) Events(w http.ResponseWriter, r *http.Request) {
	client, err := NewClient(s.hub, w, r)
	if err != nil {
		log.Printf("Failed to create WebSocket client: %v", err)
		return
	}

	s.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}

func (s *Server) Done() {
	s.hub.Done <- true
}

// func (s *SockServer) Status(w http.ResponseWriter, r *http.Request) {
// 	log.Println("STATUS")
// 	s.statusLayout.Execute(
// 		w,
// 		struct {
// 			WebsocketHost string
// 			ClientList    string
// 			PastMessages  []string
// 		}{
// 			ClientList:    s.hub.GetClientList(),
// 			WebsocketHost: "ws://" + r.Host + "/events",
// 			PastMessages:  s.PastMessages,
// 		},
// 	)
// }
