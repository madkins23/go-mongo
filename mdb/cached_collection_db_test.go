//go:build database

package mdb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/madkins23/go-type/reg"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"

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
	reg.Highlander().Clear()
	suite.Require().NoError(test.Register())
	suite.Require().NoError(test.RegisterWrapped())
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

func (suite *cacheTestSuite) TestCreateDuplicate() {
	err := suite.cached.Create(test.SimpleItem1)
	suite.Require().NoError(err)
	item, err := suite.cached.Find(test.SimpleItem1)
	suite.Require().NoError(err)
	suite.NotNil(item)
	err = suite.cached.Create(test.SimpleItem1)
	suite.Require().Error(err)
	suite.Require().True(IsDuplicate(err))
}

func (suite *cacheTestSuite) TestFindNone() {
	item, err := suite.cached.Find(test.SimpleKeyOfTheBeast)
	suite.Require().Error(err)
	suite.True(IsNotFound(err))
	suite.Nil(item)
}

func (suite *cacheTestSuite) TestFindOrCreate() {
	item, err := suite.cached.Find(test.SimpleItem2)
	suite.Require().Error(err)
	suite.True(IsNotFound(err))
	suite.Nil(item)
	item, err = suite.cached.FindOrCreate(test.SimpleItem2)
	suite.Require().NoError(err)
	suite.NotNil(item)
	item, err = suite.cached.Find(test.SimpleItem2)
	suite.Require().NoError(err)
	suite.NotNil(item)
	item2, err := suite.cached.FindOrCreate(test.SimpleItem2)
	suite.Require().NoError(err)
	suite.NotNil(item2)
	suite.Equal(item, item2)
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
	suite.True(IsNotFound(err))
	suite.Nil(noItem)
	err = suite.cached.Delete(test.SimpleItem1, false)
	suite.Require().Error(err)
	err = suite.cached.Delete(test.SimpleItem1, true)
	suite.Require().NoError(err)
}

func (suite *cacheTestSuite) TestCountDeleteAll() {
	suite.Require().NoError(suite.cached.Create(test.SimpleItem1))
	suite.Require().NoError(suite.cached.Create(test.SimpleItem2))
	suite.Require().NoError(suite.cached.Create(test.SimpleItem3))
	count, err := suite.cached.Count(bson.D{})
	suite.NoError(err)
	suite.Equal(int64(3), count)
	suite.NoError(suite.cached.DeleteAll())
	count, err = suite.cached.Count(bson.D{})
	suite.NoError(err)
	suite.Equal(int64(0), count)
}

//func (suite *cacheTestSuite) TestIterate() {
//	suite.Require().NoError(suite.cached.Create(test.SimpleItem1))
//	suite.Require().NoError(suite.cached.Create(test.SimpleItem2))
//	suite.Require().NoError(suite.cached.Create(test.SimpleItem3))
//	count := 0
//	var alpha []string
//	suite.NoError(suite.cached.Iterate(bson.D{},
//		func(item *test.SimpleItem) error {
//			alpha = append(alpha, item.Alpha)
//			count++
//			return nil
//		}))
//	suite.Equal(3, count)
//	suite.Equal([]string{"one", "two", "three"}, alpha)
//}
//
//func (suite *cacheTestSuite) TestIterateFiltered() {
//	suite.Require().NoError(suite.cached.Create(test.SimpleItem1))
//	suite.Require().NoError(suite.cached.Create(test.SimpleItem2))
//	suite.Require().NoError(suite.cached.Create(test.SimpleItem3))
//	count := 0
//	var alpha []string
//	suite.NoError(suite.cached.Iterate(bson.D{bson.E{Key: "bravo", Value: 2}},
//		func(item *test.SimpleItem) error {
//			alpha = append(alpha, item.Alpha)
//			count++
//			return nil
//		}))
//	suite.Equal(1, count)
//	suite.Equal([]string{"two"}, alpha)
//}

func (suite *cacheTestSuite) TestStringValuesFor() {
	collection, err := suite.access.Collection(context.TODO(), "mdb-cached-collection-string-values", "")
	suite.Require().NoError(err)
	cached := NewCachedCollection[*test.SimpleItem](collection, time.Hour)
	suite.NotNil(cached)
	for i := 0; i < 5; i++ {
		suite.Require().NoError(cached.Create(&test.SimpleItem{
			SimpleKey: test.SimpleKey{
				Alpha: fmt.Sprintf("Alpha #%d", i),
				Bravo: i,
			},
			Charlie: "There can be only one",
		}))
	}
	values, err := cached.StringValuesFor("alpha", nil)
	suite.Require().NoError(err)
	suite.Len(values, 5)
	values, err = cached.StringValuesFor("charlie", nil)
	suite.Require().NoError(err)
	suite.Len(values, 1)
	values, err = cached.StringValuesFor("goober", nil)
	suite.Require().NoError(err)
	suite.Len(values, 0)
}

////////////////////////////////////////////////////////////////////////////////

func (suite *cacheTestSuite) TestCreateFindDeleteWrapped() {
	collection, err := suite.access.Collection(context.TODO(), "mdb-cached-collection-wrapped-items", "")
	suite.Require().NoError(err)
	wrapped := NewCachedCollection[*WrappedItems](collection, time.Hour)
	suite.Require().NotNil(wrapped)
	suite.Require().NoError(wrapped.DeleteAll())
	wrappedItems := MakeWrappedItems()
	suite.Require().NoError(wrapped.Create(wrappedItems))
	foundWrapped, err := wrapped.Find(wrappedItems)
	suite.Require().NoError(err)
	suite.Require().NotNil(foundWrapped)
	suite.Equal(wrappedItems, foundWrapped)
	suite.Equal(test.ValueText, foundWrapped.Single.Get().String())
	for _, item := range foundWrapped.Array {
		switch item.Get().Key() {
		case "text":
			suite.Equal(test.ValueText, item.Get().String())
		case "numeric":
			if numVal, ok := item.Get().(*test.NumericValue); ok {
				suite.Equal(test.ValueNumber, numVal.Number)
			} else {
				suite.Fail("Not NumericValue: " + item.Get().String())
			}
		case "random":
			random := item.Get().String()
			fmt.Printf("Random:  %s\n", random)
			suite.True(len(random) >= test.RandomMinimum)
			suite.True(len(random) <= test.RandomMaximum)
		default:
			suite.Fail("Unknown item key: '" + item.Get().Key() + "'")
		}
	}
	for key, item := range foundWrapped.Map {
		suite.Equal(key, item.Get().Key())
	}
	cacheKey := wrappedItems.CacheKey()
	suite.NotEmpty(cacheKey)
	err = wrapped.Delete(wrappedItems, false)
	suite.Require().NoError(err)
	noItem, err := wrapped.Find(wrappedItems)
	suite.Require().Error(err)
	suite.True(IsNotFound(err))
	suite.Nil(noItem)
	err = wrapped.Delete(wrappedItems, false)
	suite.Require().Error(err)
	err = wrapped.Delete(wrappedItems, true)
	suite.Require().NoError(err)
}
