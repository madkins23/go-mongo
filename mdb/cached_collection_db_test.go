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
	cache *CachedCollection
}

func TestCacheSuite(t *testing.T) {
	suite.Run(t, new(cacheTestSuite))
}

func (suite *cacheTestSuite) SetupSuite() {
	suite.AccessTestSuite.SetupSuite()
	collection, err := suite.access.Collection(context.TODO(), "test-cache-collection", testValidatorJSON)
	suite.Require().NoError(err)
	suite.NotNil(collection)
	suite.Require().NoError(suite.access.Index(collection, NewIndexDescription(true, "alpha")))
	suite.cache = NewCachedCollection(collection, &testItem{}, time.Hour)
}

func (suite *cacheTestSuite) TearDownTest() {
	suite.NoError(suite.cache.DeleteAll())
}

func (suite *cacheTestSuite) TestFindNone() {
	item, err := suite.cache.Find(testKeyOfTheBeast)
	suite.Require().Error(err)
	suite.True(suite.cache.NotFound(err))
	suite.Nil(item)
}

func (suite *cacheTestSuite) TestCreateFindDelete() {
	err := suite.cache.Create(testItem1)
	suite.Require().NoError(err)
	item, err := suite.cache.Find(testItem1)
	suite.Require().NoError(err)
	suite.NotNil(item)
	cacheKey := testItem1.CacheKey()
	suite.NotEmpty(cacheKey)
	_, ok := suite.cache.cache[cacheKey]
	suite.True(ok)
	err = suite.cache.Delete(item, false)
	suite.Require().NoError(err)
	_, ok = suite.cache.cache[cacheKey]
	suite.False(ok)
	noItem, err := suite.cache.Find(testItem1)
	suite.Require().Error(err)
	suite.True(suite.cache.NotFound(err))
	suite.Nil(noItem)
	err = suite.cache.Delete(testItem1, false)
	suite.Require().Error(err)
	err = suite.cache.Delete(testItem1, true)
	suite.Require().NoError(err)
}

func (suite *cacheTestSuite) TestFindOrCreate() {
	item, err := suite.cache.Find(testItem2)
	suite.Require().Error(err)
	suite.True(suite.cache.NotFound(err))
	suite.Nil(item)
	item, err = suite.cache.FindOrCreate(testItem2)
	suite.Require().NoError(err)
	suite.NotNil(item)
	item, err = suite.cache.Find(testItem2)
	suite.Require().NoError(err)
	suite.NotNil(item)
	item2, err := suite.cache.FindOrCreate(testItem2)
	suite.Require().NoError(err)
	suite.NotNil(item2)
	suite.Equal(item, item2)
}
