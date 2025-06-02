package homeasst

import (
	"encoding/json"
	"fmt"
	"log"

	"gopkg.in/yaml.v2"
)

func (home *HomeRuntime) ParseResponse(buf []byte) (err error) {
	resp := &Response{}
	err = decode(buf, resp)
	if err != nil {
		return
	}

	switch resp.Type {
	case "event":
		err = home.parseEvent(buf)
	case "result":
		err = home.parseResult(buf)
	}

	return
}

func (home *HomeRuntime) parseResult(buf []byte) (err error) {
	result := &PartialResult{}
	err = decode(buf, result)
	if err != nil {
		return
	}

	if result.ID == home.loadStatesID {
		result := &StateResult{}
		err = decode(buf, result)
		if err != nil {
			log.Println(err, "parseResult")
			return
		}

		for _, entity := range result.Entities {
			home.Entities[entity.EntityID] = entity
			home.Consume(entity.EntityID, entity)
		}
		// data.loaded.Set(true)
	} else if !result.Success {
		showYaml(result)
	}
	return
}

func (home *HomeRuntime) parseEvent(buf []byte) (err error) {
	idResult := &EventResult[DataState]{}
	err = decode(buf, idResult)
	if err != nil {
		return
	}

	if idResult.Event.EventType == "state_changed" {
		entityID := idResult.Event.Data.EntityID
		result := &EventResult[DataStateChange[json.RawMessage]]{}
		err = decode(buf, result)
		if err != nil {
			return
		}

		newState := result.Event.Data.NewState
		home.Entities[entityID] = newState
		home.Consume(entityID, newState)
	}
	return
}

func decode(buf []byte, lresult any) (err error) {
	err = json.Unmarshal(buf, lresult)
	if err != nil {
		log.Println(string(buf))
		log.Println("parseAny Unmarshal", err)
	}
	return
}

func showYaml(entity any) {
	out, err := yaml.Marshal(entity)
	if err != nil {
		log.Println("Marshal yaml", err)
		return
	}
	fmt.Println(string(out))
}
