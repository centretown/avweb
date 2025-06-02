package homeasst

import (
	"fmt"
	"log"
)

type WifiSensors struct {
	Entity[SensorAttributes]
}

func (ws *WifiSensors) SignalIcon() string {
	signal := -100
	count, _ := fmt.Sscan(ws.State, &signal)
	if count == 0 {
		return "signal_wifi_bad"
	}
	if signal < -67 {
		return "signal_wifi_0_bar"
	}
	if signal < -60 {
		return "network_wifi_1_bar"
	}
	if signal < -50 {
		return "network_wifi_2_bar"
	}
	if signal < -40 {
		return "network_wifi_3_bar"
	}
	return "network_wifi"
}

func (home *HomeRuntime) WifiSensors() (sensors []*WifiSensors) {
	ids := ListEntitiesLike("wifi", home.EntityKeys)
	sensors = make([]*WifiSensors, 0, len(ids))
	for _, id := range ids {
		sensor := &WifiSensors{}
		e, ok := home.Entities[id]
		if ok {
			sensor.Copy(e)
		}
		sensors = append(sensors, sensor)
	}
	return
}

func (home *HomeRuntime) WifiSensor(entityID string) (wifi *WifiSensors) {
	wifi = &WifiSensors{}
	e, ok := home.Entities[entityID]
	if !ok {
		log.Println(entityID, "not found")
		return
	}
	wifi.Copy(e)
	return
}
