package test

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

var SimpleValidatorJSON = `{
	"$jsonSchema": {
		"bsonType": "object",
		"required": ["alpha", "bravo", "charlie"],
		"properties": {
			"alpha": {
				"bsonType": "string"
			},
			"bravo": {
				"bsonType": "int"
			},
			"charlie": {
				"bsonType": "string"
			}
		}
	}
}`

////////////////////////////////////////////////////////////////////////////////

type SimpleKey struct {
	Alpha      string
	Bravo      int
	unfiltered bool
}

func (tk *SimpleKey) CacheKey() string {
	return fmt.Sprintf("%s-%d", tk.Alpha, tk.Bravo)
}

func (tk *SimpleKey) Filter() bson.D {
	if tk.unfiltered {
		return bson.D{}
	} else {
		return bson.D{
			{"alpha", tk.Alpha},
			{"bravo", tk.Bravo},
		}
	}
}

type SimpleItem struct {
	SimpleKey `bson:"inline"`
	Charlie   string
	Delta     int
	Realized  bool
	expires   time.Time
}

func (ti *SimpleItem) ExpireAfter(duration time.Duration) {
	ti.expires = time.Now().Add(duration)
}

func (ti *SimpleItem) Expired() bool {
	return time.Now().After(ti.expires)
}

func (ti *SimpleItem) Realize() error {
	ti.Realized = true
	return nil
}

////////////////////////////////////////////////////////////////////////////////

const (
	SimpleCharlie1 = "One is the loneliest number"
)

var (
	SimpleItem1 = &SimpleItem{
		SimpleKey: SimpleKey{
			Alpha: "one",
			Bravo: 1,
		},
		Charlie: SimpleCharlie1,
		Delta:   1,
	}
	SimpleItem1x = &SimpleItem{
		SimpleKey: SimpleKey{
			Alpha: "xRay",
			Bravo: 23,
		},
		Charlie: "Do not envy the man with the x-ray eyes",
	}
	SimpleItem2 = &SimpleItem{
		SimpleKey: SimpleKey{
			Alpha: "two",
			Bravo: 2,
		},
		Charlie: "It takes two to tango",
	}
	SimpleItem3 = &SimpleItem{
		SimpleKey: SimpleKey{
			Alpha: "three",
			Bravo: 3,
		},
		Charlie: "Three can keep a secret if two of them are dead",
	}
	UnfilteredItem = &SimpleItem{
		SimpleKey: SimpleKey{
			Alpha:      "Camel",
			Bravo:      100,
			unfiltered: true,
		},
		Charlie: "Mirage",
	}
	SimplyInvalid = &SimpleKey{
		Alpha: "beast",
		Bravo: 666,
	}
)
