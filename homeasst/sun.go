package homeasst

import (
	"slices"
	"strings"
	"time"

	"github.com/centretown/avweb/action"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const ShortTime = "2006-01-02 15:04:05"

type TimeStampSensor struct {
	Sun       Entity[TimeStampAttributes]
	TimeStamp time.Time
	Action    *action.Action
}

func (tss *TimeStampSensor) ShortName() string {
	c := cases.Title(language.Und, cases.NoLower)
	return c.String(strings.
		TrimPrefix(tss.Sun.Attributes.Name, "Sun Next "))
}

func (tss *TimeStampSensor) FormatTime() string {
	return tss.TimeStamp.Local().Format(ShortTime)
}

type Sun struct {
	Action  *action.Action
	Sensors []*TimeStampSensor
}

func (home *HomeRuntime) NewSun(action *action.Action) (sun *Sun) {
	sunlist := ListEntitiesLike("sensor.sun_next", home.EntityKeys)
	sensors := make([]*TimeStampSensor, 0, len(sunlist))
	for _, s := range sunlist {
		sensor := &TimeStampSensor{}
		sensor.Sun.Copy(home.Entities[s])
		sensor.TimeStamp, _ = time.Parse(time.RFC3339, sensor.Sun.State)
		sensors = append(sensors, sensor)
	}
	slices.SortFunc(sensors, func(a, b *TimeStampSensor) int {
		return a.TimeStamp.Compare(b.TimeStamp)
	})
	sun = &Sun{Action: action, Sensors: sensors}
	return
}
