package homeasst

import (
	"testing"
)

func TestAppData(t *testing.T) {
	data, err := NewHomeRuntime()
	if err != nil {
		t.Fatal(err)
	}
	testAuthorize(t, data)
	testBuildEntities(t, data)
	testPrefixWifi(t, data)
	testSunTimes(t, data)
}

func testAuthorize(t *testing.T, data *HomeRuntime) {
	ok, err := data.Authorize()
	if err != nil {
		t.Fatal("authorize", err)
	}
	if !ok {
		t.Fatal("not authorized")

	}
	t.Log("authorized")
}

func testSunTimes(t *testing.T, data *HomeRuntime) {
	// sensors := data.SunTimes(nil)
	// tmpl, err := template.ParseGlob("../www/*.html")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// err = tmpl.Lookup("layout.sun").Execute(os.Stderr, sensors)
	// if err != nil {
	// 	t.Fatal(err)
	// }
}

func testPrefixWifi(t *testing.T, data *HomeRuntime) {
	show := func(list []string) {
		for i := range list {
			t.Log(list[i])
		}
	}
	show(ListEntitiesLike("sensor.wifi", data.EntityKeys))
	show(ListEntitiesLike("sensor.sun_next", data.EntityKeys))
}

func testBuildEntities(t *testing.T, data *HomeRuntime) {
	var err error = data.BuildEntities()
	if err != nil {
		t.Fatal("BuildEntities", err)
	}

	var (
		weather Entity[WeatherAttributes]
		wifi    Entity[SensorAttributes]
	)

	for k, v := range data.Entities {
		switch k {
		case "weather.forecast_home":
			weather.Copy(v)
			showYaml(&weather)
		case "sensor.wifi_signal_30":
			wifi.Copy(v)
			showYaml(&wifi)
		}
		t.Log(k)
	}

}

// func test_subscribe(t *testing.T, data *HomeData) {
// 	var (
// 		brightValue      int
// 		effect           string
// 		red, green, blue uint8
// 	)

// 	lightIDs := []string{
// 		"light.led_matrix_24",
// 		"light.led_strip_24"}

// 	for _, id := range lightIDs {
// 		var lp Light
// 		sub := NewSubcription(&lp.Entity, func(c Consumer) {
// 			if lp.Attributes.Brightness != brightValue {
// 				brightValue = lp.Attributes.Brightness
// 				v := float64(brightValue) * 100 / 255
// 				log.Println(v)
// 			}
// 			if lp.Attributes.Effect != effect {
// 				effect = lp.Attributes.Effect
// 				log.Println(effect)
// 			}

// 			rgb := lp.Attributes.ColorRGB
// 			if len(rgb) > 2 {
// 				if red != rgb[0] || green != rgb[1] || blue != rgb[2] {
// 					red, green, blue = rgb[0], rgb[1], rgb[2]
// 					t.Log("red", red, "green", green, "blue", blue)
// 				}
// 			}
// 		})
// 		data.Subscribe(id, sub)
// 	}

// }
