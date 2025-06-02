package homeasst

import "encoding/json"

type Consumer interface {
	Copy(src *Entity[json.RawMessage])
}

type Subscription struct {
	consumer Consumer
	Run      func(consumer Consumer)
	Enabled  bool
}

func NewSubcription(consumer Consumer, run func(Consumer)) *Subscription {
	sub := &Subscription{
		consumer: consumer,
		Run:      run,
		Enabled:  true,
	}
	return sub
}

func (sub *Subscription) Consume(newState *Entity[json.RawMessage]) {
	sub.consumer.Copy(newState)
	sub.Run(sub.consumer)
}
