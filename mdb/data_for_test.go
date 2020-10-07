package mdb

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

var testValidatorJSON = `{
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

type TestKey struct {
	Alpha string
	Bravo int
}

func (tk *TestKey) CacheKey() string {
	return fmt.Sprintf("%s-%d", tk.Alpha, tk.Bravo)
}

func (tk *TestKey) Filter() bson.D {
	return bson.D{
		{"alpha", tk.Alpha},
		{"bravo", tk.Bravo},
	}
}

type testItem struct {
	TestKey `bson:"inline"`
	Charlie string
	expires time.Time
}

func (ti *testItem) ExpireAfter(duration time.Duration) {
	ti.expires = time.Now().Add(duration)
}

func (ti *testItem) Expired() bool {
	return time.Now().After(ti.expires)
}

func (ti *testItem) Filter() bson.D {
	return bson.D{
		{"alpha", ti.Alpha},
		{"bravo", ti.Bravo},
	}
}

func (ti *testItem) Realize() error {
	return nil
}

////////////////////////////////////////////////////////////////////////////////

var (
	testItem1 = &testItem{
		TestKey: TestKey{
			Alpha: "one",
			Bravo: 1,
		},
		Charlie: "One is the loneliest number",
	}
	testItem2 = &testItem{
		TestKey: TestKey{
			Alpha: "two",
			Bravo: 2,
		},
		Charlie: "It takes two to tango",
	}
	testItem3 = &testItem{
		TestKey: TestKey{
			Alpha: "three",
			Bravo: 3,
		},
		Charlie: "Three can keep a secret if two of them are dead",
	}
	testKeyOfTheBeast = &TestKey{
		Alpha: "beast",
		Bravo: 666,
	}
)
