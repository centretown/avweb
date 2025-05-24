package runtime

import (
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var schema = `
CREATE TABLE IF NOT EXISTS location
(ID INTEGER PRIMARY KEY,
City TEXT,
Latitude REAL,
Longitude REAL,
Zone TEXT);

CREATE UNIQUE INDEX IF NOT EXISTS location_index
ON location (City, Latitude, Longitude);

CREATE TABLE IF NOT EXISTS history
(LocationID INTEGER,
Time TEXT,
Interval REAL,
Temperature REAL,
Precipitation REAL,
Humidity REAL,
FeelsLike REAL,
IsDay INTEGER,
Code INTEGER,
WindSpeed REAL,
WindDirection REAL,
WindGusts REAL,
Rain REAL,
Showers REAL,
Snowfall REAL,
CloudCover REAL,
PressureMSL REAL,
SurfacePressure REAL,
PRIMARY KEY (LocationID, Time) );
`

var insertLocation = `INSERT OR IGNORE INTO location (City,
Latitude,
Longitude,
Zone)
VALUES (?, ?, ?, ?);`

var insertHistory = `INSERT OR IGNORE INTO history (
LocationID,
Time,
Interval,
Temperature,
Precipitation,
Humidity,
FeelsLike,
IsDay,
Code,
WindSpeed,
WindDirection,
WindGusts,
Rain,
Showers,
Snowfall,
CloudCover,
PressureMSL,
SurfacePressure)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

func OpenDB(filename string) (db *sqlx.DB, err error) {
	db, err = sqlx.Connect("sqlite3", filename)
	if err != nil {
		return
	}
	_, err = db.Exec(schema)
	return
}
func InsertLocation(db *sqlx.DB, location *Location) (err error) {
	_, err = db.Exec(insertLocation,
		location.City,
		location.Latitude,
		location.Longitude,
		location.Zone)
	return
}

func InsertHistory(db *sqlx.DB, ID uint64, current *Current) (err error) {
	_, err = db.Exec(insertHistory, ID,
		current.Time,
		current.Interval,
		current.Temperature,
		current.Precipitation,
		current.Humidity,
		current.FeelsLike,
		current.IsDay,
		current.Code,
		current.WindSpeed,
		current.WindDirection,
		current.WindGusts,
		current.Rain,
		current.Showers,
		current.Snowfall,
		current.CloudCover,
		current.PressureMSL,
		current.SurfacePressure)
	return
}

const (
	fmtSelectHistory   = "SELECT * FROM history WHERE Time>'%s' AND Time<='%s' ORDER BY LocationID, Time;"
	fmtLocationHistory = "SELECT * FROM history WHERE LocationID=%d AND Time>'%s' AND Time<='%s' ORDER BY Time %s;"
	timeLayout         = "2006-01-02T15:04"
	fmtSelectLocation  = "SELECT * FROM location;"
)

func SelectHistoryInterval(db *sqlx.DB, ID uint64, after string, before string, order string) (history []*Current, err error) {
	query := fmt.Sprintf(fmtLocationHistory, ID, after, before, order)
	return SelectHistory(db, query)
}

func SelectHistory(db *sqlx.DB, query string) (history []*Current, err error) {
	var rows *sqlx.Rows
	rows, err = db.Queryx(query)
	if err != nil {
		log.Println(err)
		return
	}

	for rows.Next() {
		var current = &Current{}
		err = rows.StructScan(current)
		if err != nil {
			log.Println(err)
			return
		} else {
			history = append(history, current)
		}
	}
	return
}

func SelectLocations(db *sqlx.DB) (locations []*Location, err error) {
	var rows *sqlx.Rows
	rows, err = db.Queryx(fmtSelectLocation)
	if err != nil {
		log.Println(err)
		return
	}

	for rows.Next() {
		var location = &Location{}
		err = rows.StructScan(location)
		if err != nil {
			log.Println(err)
			return
		} else {
			locations = append(locations, location)
		}
	}
	return
}

func BeforeTime(t time.Time, d time.Duration) (after string, before string) {
	before = t.Format(timeLayout)
	after = t.Add(-d).Format(timeLayout)
	return
}
