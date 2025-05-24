package runtime

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestCreateDB(t *testing.T) {
	t.Log("TestDB!")
	db, err := OpenDB("testdata/location.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	t.Log("db opened!")

	config, err := testLoadConfig(t)
	if err != nil {
		t.Fatal(err)
	}
	for _, location := range config.Locations {
		err = InsertLocation(db, location)
		if err != nil {
			t.Fatal(err)
		}
	}

	list, err := SelectLocations(db)
	if err != nil {
		t.Fatal(err)
	}
	for i, location := range list {
		t.Log(i, location)
	}
	// return

	history, err := testLoadHistory(t)
	if err != nil {
		t.Fatal(err)
	}
	for i, area := range history {
		location := list[i]
		for _, current := range area {
			err = InsertHistory(db, location.ID, current)
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	listH, err := SelectHistoryInterval(db, 1, "2025-05-21T23:45", "2025-05-22T02:00", "ASC")
	if err != nil {
		t.Fatal(err)
	}
	for i, area := range listH {
		t.Log(i, area)
	}
}

func testLoadConfig(t *testing.T) (config *Config, err error) {
	config = &Config{}
	err = config.Read("testdata/config.json")
	if err != nil {
		t.Fatal(err)
	}
	return
}

func testLoadHistory(t *testing.T) (history [][]*Current, err error) {
	var buf []byte
	history = make([][]*Current, 0)
	buf, err = os.ReadFile("testdata/history.json")
	if err != nil {
		t.Log(err)
		return
	}
	err = json.Unmarshal(buf, &history)
	if err != nil {
		t.Log(err)
		return
	}
	return
}

func TestQueryDB(t *testing.T) {
	t.Log("TestQueryDB")
	db, err := OpenDB("testdata/location.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	t.Log("opened")

	after, before := BeforeTime(time.Now(), 8*time.Hour)
	historyList, err := SelectHistoryInterval(db, 1, after, before, "ASC")
	if err != nil {
		t.Fatal(err)
	}
	for i, currentItem := range historyList {
		if !(currentItem.Time > after && currentItem.Time <= before) {
			t.Fatal("!(area.Time > after && area.Time <= before)", currentItem.Time)
		}
		t.Log(i, currentItem)
	}
}
