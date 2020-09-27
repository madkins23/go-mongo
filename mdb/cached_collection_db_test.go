// +build database

package mdb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
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
	suite.collection, err = suite.access.Collection("test-cache-collection", testValidatorJSON)
	suite.Require().NoError(err)
	suite.NotNil(suite.collection)
	suite.Require().NoError(suite.access.Index(suite.collection, NewIndexDescription(true, "alpha")))
	suite.cache = NewCachedCollection(suite.access, suite.collection, context.TODO(), &testItem{}, time.Hour)
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

func (suite *cacheTestSuite) TestCreateDuplicate() {
	tk := &testKey{
		alpha: "two",
		bravo: 2,
	}
	ti := &testItem{
		testKey: *tk,
		charlie: "Two is too much",
	}
	err := suite.cache.Create(ti)
	suite.Require().NoError(err)
	item, err := suite.cache.Find(tk)
	suite.Require().NoError(err)
	suite.NotNil(item)
	err = suite.cache.Create(ti)
	suite.Require().Error(err)
	suite.Require().True(suite.access.Duplicate(err))
}

func (suite *cacheTestSuite) TestFindOrCreate() {
	tk := &testKey{
		alpha: "three",
		bravo: 3,
	}
	ti := &testItem{
		testKey: *tk,
		charlie: "Three can keep a secret, if two of them are dead",
	}
	item, err := suite.cache.Find(tk)
	suite.Require().Error(err)
	suite.True(suite.cache.NotFound(err))
	suite.Nil(item)
	item, err = suite.cache.FindOrCreate(ti)
	suite.Require().NoError(err)
	suite.NotNil(item)
	item, err = suite.cache.Find(tk)
	suite.Require().NoError(err)
	suite.NotNil(item)
	item2, err := suite.cache.FindOrCreate(ti)
	suite.Require().NoError(err)
	suite.NotNil(item2)
	suite.Equal(item, item2)
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
