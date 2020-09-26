package mdb

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type cacheTestSuite struct {
	AccessTestSuite
	cache      *CachedCollection
	collection *mongo.Collection
}

func TestCacheSuite(t *testing.T) {
	suite.Run(t, new(cacheTestSuite))
}

func (suite *cacheTestSuite) SetupSuite() {
	var err error
	suite.AccessTestSuite.SetupSuite()
	suite.collection, err = suite.access.Collection("test-cache-collection", testCacheValidatorJSON)
	suite.Require().NoError(err)
	suite.NotNil(suite.collection)
	suite.cache = NewCache(suite.access, suite.collection, context.TODO(), &testItem{}, time.Hour)
}

func (suite *cacheTestSuite) TestCreateFindDelete() {
	tk := &testKey{
		alpha: "one",
		bravo: 1,
	}
	ti := &testItem{
		testKey: *tk,
		charlie: "One is the loneliest number",
	}
	err := suite.cache.Create(ti)
	suite.Require().NoError(err)
	item, err := suite.cache.Find(tk)
	suite.Require().NoError(err)
	suite.NotNil(item)
	err = suite.cache.Delete(item, false)
	suite.Require().NoError(err)
	noItem, err := suite.cache.Find(tk)
	suite.Require().Error(err)
	suite.True(suite.cache.NotFound(err))
	suite.Nil(noItem)
	err = suite.cache.Delete(tk, false)
	suite.Require().Error(err)
	err = suite.cache.Delete(tk, true)
	suite.Require().NoError(err)
}

func (suite *cacheTestSuite) TestFindNone() {
	tk := &testKey{
		alpha: "beast",
		bravo: 666,
	}
	item, err := suite.cache.Find(tk)
	suite.Require().Error(err)
	suite.True(suite.cache.NotFound(err))
	suite.Nil(item)
}

func (suite *cacheTestSuite) TestCacheKey() {
	tk := &testKey{
		bravo: 666,
	}
	cacheKey, err := tk.CacheKey()
	suite.Require().Error(err)
	suite.Equal("", cacheKey)
	_, err = suite.cache.Find(tk)
	suite.Require().Error(err)
}

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

var testCacheValidatorJSON = `{
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
