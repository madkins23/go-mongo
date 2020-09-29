// +build database

package mdb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type cacheTestSuite struct {
	AccessTestSuite
	cache      *CachedCollection
	collection *Collection
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
	suite.cache = NewCachedCollection(suite.collection, context.TODO(), &testItem{}, time.Hour)
}

func (suite *cacheTestSuite) TestFindNone() {
	tk := &TestKey{
		Alpha: "beast",
		Bravo: 666,
	}
	item, err := suite.cache.Find(tk)
	suite.Require().Error(err)
	suite.True(suite.cache.NotFound(err))
	suite.Nil(item)
}

func (suite *cacheTestSuite) TestCreateFindDelete() {
	tk := &TestKey{
		Alpha: "one",
		Bravo: 1,
	}
	ti := &testItem{
		TestKey: *tk,
		Charlie: "One is the loneliest number",
	}
	err := suite.cache.Create(ti)
	suite.Require().NoError(err)
	item, err := suite.cache.Find(tk)
	suite.Require().NoError(err)
	suite.NotNil(item)
	cacheKey := tk.CacheKey()
	suite.NotEmpty(cacheKey)
	_, ok := suite.cache.cache[cacheKey]
	suite.True(ok)
	err = suite.cache.Delete(item, false)
	suite.Require().NoError(err)
	_, ok = suite.cache.cache[cacheKey]
	suite.False(ok)
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
	tk := &TestKey{
		Alpha: "two",
		Bravo: 2,
	}
	ti := &testItem{
		TestKey: *tk,
		Charlie: "Two is too much",
	}
	err := suite.cache.Create(ti)
	suite.Require().NoError(err)
	item, err := suite.cache.Find(tk)
	suite.Require().NoError(err)
	suite.NotNil(item)
	err = suite.cache.Create(ti)
	suite.Require().Error(err)
	suite.Require().True(suite.access.Duplicate(err))
	cacheKey := tk.CacheKey()
	_, ok := suite.cache.cache[cacheKey]
	suite.True(ok)
	suite.cache.InvalidateByPrefix("two")
	_, ok = suite.cache.cache[cacheKey]
	suite.False(ok)
}

func (suite *cacheTestSuite) TestFindOrCreate() {
	tk := &TestKey{
		Alpha: "three",
		Bravo: 3,
	}
	ti := &testItem{
		TestKey: *tk,
		Charlie: "Three can keep a secret, if two of them are dead",
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
