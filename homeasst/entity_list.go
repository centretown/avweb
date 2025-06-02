package homeasst

import (
	"slices"
	"strings"
)

// func init() {
// 	slices.Sort(entityKeys)
// }

// var entityKeys = []string{
// 	"sensor.sm_a037w_battery_level",
// 	"number.tilt_2",
// 	"sensor.connected_clients",
// 	"sensor.sun_next_dawn",
// 	"binary_sensor.gimbal_toggle_button_2",
// 	"sensor.sun_next_rising",
// 	"device_tracker.sm_a037w",
// 	"sensor.sm_a037w_battery_state",
// 	"device_tracker.huawei_mla_l03",
// 	"light.led_matrix_24",
// 	"zone.home",
// 	"sensor.sun_next_midnight",
// 	"sensor.sun_next_noon",
// 	"light.led_strip_24",
// 	"light.status",
// 	"conversation.home_assistant",
// 	"sensor.sun_next_setting",
// 	"tts.google_en_com",
// 	"switch.gimbal_toggle_switch_2",
// 	"sensor.wifi_signal_30",
// 	"sensor.sm_a037w_charger_type",
// 	"sensor.huawei_mla_l03_battery_state",
// 	"weather.forecast_home",
// 	"sensor.doorstop_wifi",
// 	"person.dave",
// 	"sensor.huawei_mla_l03_battery_level",
// 	"sensor.gimbal_rotary_2",
// 	"camera.cam30",
// 	"sensor.wifi_signal_24",
// 	"sensor.ip_address_24",
// 	"number.pan_2",
// 	"todo.shopping_list",
// 	"sun.sun",
// 	"sensor.sun_next_dusk",
// }

func ListEntitiesLike(prefix string, sortedKeys []string) (list []string) {
	compare := func(s string) bool {
		return strings.Contains(s, prefix)
	}

	index := slices.IndexFunc(sortedKeys, compare)
	if index == -1 {
		return
	}

	for i := index; i < len(sortedKeys); i++ {
		s := sortedKeys[i]
		if compare(s) {
			list = append(list, s)
		}
	}
	return
}

func BuildEntityKeys(entities EntityMap) (keys []string) {
	keys = make([]string, 0, len(entities))
	for k := range entities {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return
}
