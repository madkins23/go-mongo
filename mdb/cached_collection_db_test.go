//go:build database

package mdb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/madkins23/go-mongo/test"
)

type cacheTestSuite struct {
	AccessTestSuite
	cached *CachedCollection[*test.SimpleItem]
}

func TestCacheSuite(t *testing.T) {
	suite.Run(t, new(cacheTestSuite))
}

func (suite *cacheTestSuite) SetupSuite() {
	suite.AccessTestSuite.SetupSuite()
	collection, err := suite.access.Collection(context.TODO(), "test-cache-collection", test.SimpleValidatorJSON)
	suite.Require().NoError(err)
	suite.NotNil(collection)
	suite.Require().NoError(suite.access.Index(collection, NewIndexDescription(true, "alpha")))
	suite.cached = NewCachedCollection[*test.SimpleItem](collection, time.Hour)
	suite.Require().NoError(suite.cached.DeleteAll())
}

func (suite *cacheTestSuite) TearDownTest() {
	suite.NoError(suite.cached.DeleteAll())
}

func (suite *cacheTestSuite) TestFindNone() {
	item, err := suite.cached.Find(test.SimpleKeyOfTheBeast)
	suite.Require().Error(err)
	suite.True(suite.cached.IsNotFound(err))
	suite.Nil(item)
}

func (suite *cacheTestSuite) TestCreateFindDelete() {
	err := suite.cached.Create(test.SimpleItem1)
	suite.Require().NoError(err)
	item, err := suite.cached.Find(test.SimpleItem1)
	suite.Require().NoError(err)
	suite.NotNil(item)
	cacheKey := test.SimpleItem1.CacheKey()
	suite.NotEmpty(cacheKey)
	_, ok := suite.cached.cache[cacheKey]
	suite.True(ok)
	err = suite.cached.Delete(item, false)
	suite.Require().NoError(err)
	_, ok = suite.cached.cache[cacheKey]
	suite.False(ok)
	noItem, err := suite.cached.Find(test.SimpleItem1)
	suite.Require().Error(err)
	suite.True(suite.cached.IsNotFound(err))
	suite.Nil(noItem)
	err = suite.cached.Delete(test.SimpleItem1, false)
	suite.Require().Error(err)
	err = suite.cached.Delete(test.SimpleItem1, true)
	suite.Require().NoError(err)
}

//func (suite *cacheTestSuite) TestFindOrCreate() {
//	item, err := suite.cache.Find(test.SimpleItem2)
//	suite.Require().Error(err)
//	suite.True(suite.cache.IsNotFound(err))
//	suite.Nil(item)
//	item, err = suite.cache.FindOrCreate(test.SimpleItem2)
//	suite.Require().NoError(err)
//	suite.NotNil(item)
//	ti, ok := item.(*test.SimpleItem)
//	suite.Require().True(ok)
//	suite.True(ti.Realized)
//	item, err = suite.cache.Find(test.SimpleItem2)
//	suite.Require().NoError(err)
//	suite.NotNil(item)
//	ti, ok = item.(*test.SimpleItem)
//	suite.Require().True(ok)
//	suite.True(ti.Realized)
//	item2, err := suite.cache.FindOrCreate(test.SimpleItem2)
//	suite.Require().NoError(err)
//	suite.NotNil(item2)
//	suite.Equal(item, item2)
//}

////////////////////////////////////////////////////////////////////////////////
//
//type CacheableItem struct {
//	test.SimpleItem
//	expire time.Time
//}
//
//func (ci *CacheableItem) ExpireAfter(duration time.Duration) {
//	ci.expire = time.Now().Add(duration)
//}
//
//func (ci *CacheableItem) Expired() bool {
//	return time.Now().After(ci.expire)
//}
