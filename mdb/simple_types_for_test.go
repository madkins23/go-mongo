package mdb

import (
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

var _ Identifier = &SimpleItem{}

type SimpleItem struct {
	Identity `bson:"inline"`
	Alpha    string `bson:",omitempty"`
	Bravo    int    `bson:",omitempty"`
	Charlie  string `bson:",omitempty"`
	Delta    int    `bson:",omitempty"`
	expires  time.Time
}

func (si *SimpleItem) ExpireAfter(duration time.Duration) {
	si.expires = time.Now().Add(duration)
}

func (si *SimpleItem) Expired() bool {
	return time.Now().After(si.expires)
}

// Filter returns a filter for the alpha/bravo of this item.
func (si *SimpleItem) Filter() bson.D {
	return bson.D{
		{"alpha", si.Alpha},
		{"bravo", si.Bravo},
	}
}

////////////////////////////////////////////////////////////////////////////////

const (
	SimpleCharlie1 = "One is the loneliest number"
)

var (
	SimpleItem1 = &SimpleItem{
		Alpha:   "one",
		Bravo:   1,
		Charlie: SimpleCharlie1,
		Delta:   1,
	}
	SimpleItem1x = &SimpleItem{
		Alpha:   "xRay",
		Bravo:   23,
		Charlie: "Do not envy the man with the x-ray eyes",
	}
	SimpleItem2 = &SimpleItem{
		Alpha:   "two",
		Bravo:   2,
		Charlie: "It takes two to tango",
	}
	SimpleItem3 = &SimpleItem{
		Alpha:   "three",
		Bravo:   3,
		Charlie: "Three can keep a secret if two of them are dead",
	}
	UnfilteredItem = &SimpleItem{
		Alpha:   "Camel",
		Bravo:   100,
		Charlie: "Mirage",
	}
	SimplyInvalid = &SimpleItem{
		Alpha: "Invalid",
		// Missing bravo and charlie which are required by the JSON validation above.
	}
)
