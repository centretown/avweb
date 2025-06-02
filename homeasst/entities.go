package homeasst

import (
	"encoding/json"
	"time"
)

type AuthResult struct {
	Type    string `json:"type" yaml:"type"`
	Version string `json:"ha_version" yaml:"ha_version"`
}

type Response struct {
	ID   int    `json:"id" yaml:"id"`
	Type string `json:"type" yaml:"type"`
}

type PartialResult struct {
	Response
	Success bool `json:"success" yaml:"success"`
}

type StateResult struct {
	Response
	Success  bool                       `json:"success" yaml:"success"`
	Entities []*Entity[json.RawMessage] `json:"result" yaml:"result"`
}

type DataState struct {
	EntityID string `json:"entity_id" yaml:"entity_id"`
}

type DataStateChange[T any] struct {
	EntityID string     `json:"entity_id" yaml:"entity_id"`
	OldState *Entity[T] `json:"old_state" yaml:"old_state"`
	NewState *Entity[T] `json:"new_state" yaml:"new_state"`
}

type Context struct {
	ID       string `json:"id" yaml:"id"`
	ParentID string `json:"parent_id" yaml:"parent_id"`
	UserID   string `json:"user_id" yaml:"user_id"`
}

type Event[T any] struct {
	EventType string    `json:"event_type" yaml:"event_type"`
	Origin    string    `json:"origin" yaml:"origin"`
	TimeFired time.Time `json:"time_fired" yaml:"time_fired"`
	Context   Context   `json:"context" yaml:"context"`
	Data      T         `json:"data" yaml:"data"`
}

type EventResult[T any] struct {
	Response
	Event Event[T] `json:"event" yaml:"event"`
}

type LightEventResult struct {
	EventResult[LightAttributes]
}
