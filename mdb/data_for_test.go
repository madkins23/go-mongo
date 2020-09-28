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

type TestKey struct {
	Alpha string
	Bravo int
}

func (tk *TestKey) CacheKey() (string, error) {
	return fmt.Sprintf("%s-%d", tk.Alpha, tk.Bravo), nil
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

func (ti *testItem) Document() bson.M {
	return bson.M{
		"alpha":   ti.TestKey.Alpha,
		"bravo":   ti.TestKey.Bravo,
		"charlie": ti.Charlie,
	}
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
