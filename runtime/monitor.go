package runtime

import (
	"log"
	"time"
)

func FirstTicker() (ticker time.Duration) {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(),
		now.Minute(), 0, 0, now.Location())
	minute := 15 - next.Minute()%15
	if minute == 0 {
		minute = 15
	}
	next = next.Add(time.Duration(minute) * time.Minute)
	if next.Compare(now) < 0 {
		log.Fatal(minute, now, next)
	}
	ticker = next.Sub(now)
	log.Printf("ticker=%v, minute=%v now=%v next=%v\n", ticker, minute, now, next)
	return
}

func (rt *Runtime) Monitor() {
	log.Println("start Monitoring")
	initialRun := true
	for {
		select {
		case now := <-rt.ticker.C:
			if initialRun {
				rt.ticker.Reset(time.Minute * 15)
				initialRun = false
			}
			log.Println("QueryCurrent")
			rt.QueryCurrent()
			if now.Minute() == 0 {
				log.Println("QueryHourly")
				rt.QueryHourly()
				if now.Hour()%4 == 0 {
					log.Println("QueryDaily")
					rt.QueryDaily()
				}
			}
			rt.BroadcastTemperature()
		}

		time.Sleep(time.Second)
	}
}
