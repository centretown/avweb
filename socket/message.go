package socket

import "time"

const maxMessages = 10

type Message struct {
	Name    string
	Message string
	Stamp   time.Time
}

func (msg *Message) StampShort() string {
	return msg.Stamp.Local().Format(time.TimeOnly)
}
