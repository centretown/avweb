package homeasst

import (
	"fmt"
)

type LightAttributes struct {
	Name       string    `json:"friendly_name" yaml:"friendly_name"`
	Brightness int       `json:"brightness" yaml:"brightness"`
	ColorMode  string    `json:"color_mode" yaml:"color_mode"`
	Effect     string    `json:"effect" yaml:"effect"`
	EffectList []string  `json:"effect_list" yaml:"effect_list"`
	ColorRGB   []uint8   `json:"rgb_color" yaml:"rgb_color"`
	ColorXY    []float64 `json:"xy_color" yaml:"xy_color"`
	ColorHS    []float64 `json:"hs_color" yaml:"hs_color"`
}

type Light struct {
	Entity[LightAttributes]
}

func (led *Light) HexColor() string {
	if len(led.Attributes.ColorRGB) >= 3 {
		return fmt.Sprintf("#%02x%02x%02x",
			led.Attributes.ColorRGB[0],
			led.Attributes.ColorRGB[1],
			led.Attributes.ColorRGB[2])
	}
	return "#3f3f3f"
}

func (home *HomeRuntime) NewLedLights() (lights []*Light) {
	ids := ListEntitiesLike("light.led", home.EntityKeys)
	lights = make([]*Light, 0, len(ids))
	for _, id := range ids {
		light := &Light{}
		e, ok := home.Entities[id]
		if ok {
			light.Copy(e)
		}
		lights = append(lights, light)
	}
	return
}
