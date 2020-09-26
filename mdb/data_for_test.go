package mdb

import (
	"errors"
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

var (
	errAlphaNotString   = errors.New("alpha not a string")
	errBravoNotInt      = errors.New("bravo not an int")
	errCharlieNotString = errors.New("charlie not a string")
	errNoFieldAlpha     = errors.New("no alpha")
	errNoFieldBravo     = errors.New("no bravo")
)

type testKey struct {
	alpha string
	bravo int
}

func (tk *testKey) CacheKey() (string, error) {
	if tk.alpha == "" {
		return "", errAlphaNotString
	}

	return fmt.Sprintf("%s-%d", tk.alpha, tk.bravo), nil
}

func (tk *testKey) Filter() bson.D {
	return bson.D{
		{"alpha", tk.alpha},
		{"bravo", tk.bravo},
	}
}

type testItem struct {
	testKey
	charlie string
	expires time.Time
}

func (ti *testItem) Document() bson.M {
	return bson.M{
		"alpha":   ti.alpha,
		"bravo":   ti.bravo,
		"charlie": ti.charlie,
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
		{"alpha", ti.alpha},
		{"bravo", ti.bravo},
	}
}

func (ti *testItem) InitFrom(stub bson.M) error {
	var ok bool
	if alpha, found := stub["alpha"]; !found {
		return errNoFieldAlpha
	} else if ti.alpha, ok = alpha.(string); !ok {
		return errAlphaNotString
	}
	if bravo, found := stub["bravo"]; !found {
		return errNoFieldBravo
	} else if bravo32, ok := bravo.(int32); !ok {
		return errBravoNotInt
	} else {
		ti.bravo = int(bravo32)
	}
	if charlie, found := stub["charlie"]; !found {
		// Field not required.
		ti.charlie = ""
	} else if ti.charlie, ok = charlie.(string); !ok {
		return errCharlieNotString
	}
	return nil
}
