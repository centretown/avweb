package homeasst

// func NewNumberList(entities []string) []*Number {
// 	nums := make([]*Number, 0, len(entities))
// 	for _, ent := range entities {
// 		buf, err := websock.Get("states/" + ent)
// 		if err != nil {
// 			log.Println(err, string(buf))
// 			continue
// 		}

// 		num := &Number{}
// 		err = json.Unmarshal(buf, num)
// 		if err != nil {
// 			log.Println(err, string(buf))
// 			continue
// 		}
// 		nums = append(nums, num)
// 	}
// 	return nums
// }

/*
{
    "entity_id": "number.tilt",
    "state": "162.0",
    "attributes": {
        "min": 0.0,
        "max": 180.0,
        "step": 1.0,
        "mode": "slider",
        "unit_of_measurement": "°",
        "friendly_name": "Tilt"
    },
    "last_changed": "2024-06-28T01:52:05.073625+00:00",
    "last_reported": "2024-06-28T01:52:05.073625+00:00",
    "last_updated": "2024-06-28T01:52:05.073625+00:00",
    "context": {
        "id": "01J1E8NSBZCFFGP93ERQ0DRW0W",
        "parent_id": null,
        "user_id": "4fcc5ee6683d4c9eb1ac5e8e1e42d240"
    }
}
{
    "entity_id": "number.pan",
    "state": "90.0",
    "attributes": {
        "min": 0.0,
        "max": 180.0,
        "step": 1.0,
        "mode": "slider",
        "unit_of_measurement": "°",
        "friendly_name": "Pan"
    },
    "last_changed": "2024-06-28T22:35:00.655655+00:00",
    "last_reported": "2024-06-28T22:35:00.655655+00:00",
    "last_updated": "2024-06-28T22:35:00.655655+00:00",
    "context": {
        "id": "01J1GFSN2YN04X3RKCJEG4AKZM",
        "parent_id": null,
        "user_id": "4fcc5ee6683d4c9eb1ac5e8e1e42d240"
    }
}*/
