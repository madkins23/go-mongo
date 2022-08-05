//go:build database

package mdb

import (
	"context"
	"fmt"
	"testing"

	"github.com/madkins23/go-type/reg"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/madkins23/go-mongo/test"
)

type typedTestSuite struct {
	AccessTestSuite
	typed *TypedCollection[test.SimpleItem]
}

func TestTypedSuite(t *testing.T) {
	suite.Run(t, new(typedTestSuite))
}

func (suite *typedTestSuite) SetupSuite() {
	suite.AccessTestSuite.SetupSuite()
	reg.Highlander().Clear()
	suite.Require().NoError(test.Register())
	suite.Require().NoError(test.RegisterWrapped())
	collection, err := suite.access.Collection(context.TODO(), "test-collection", test.SimpleValidatorJSON)
	suite.Require().NoError(err)
	suite.NotNil(collection)
	suite.Require().NoError(suite.access.Index(collection, NewIndexDescription(true, "alpha")))
	suite.typed = NewTypedCollection[test.SimpleItem](collection)
	suite.Require().NoError(suite.typed.DeleteAll())
	suite.Require().NotNil(suite.typed)
}

func (suite *typedTestSuite) TearDownTest() {
	suite.NoError(suite.typed.DeleteAll())
}

func (suite *typedTestSuite) TestCreateDuplicate() {
	err := suite.typed.Create(test.SimpleItem1)
	suite.Require().NoError(err)
	item, err := suite.typed.Find(test.SimpleItem1.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	err = suite.typed.Create(test.SimpleItem1)
	suite.Require().Error(err)
	suite.Require().True(IsDuplicate(err))
}

func (suite *typedTestSuite) TestFindNone() {
	item, err := suite.typed.Find(test.SimpleKeyOfTheBeast.Filter())
	suite.Require().Error(err)
	suite.True(IsNotFound(err))
	suite.Nil(item)
}

func (suite *typedTestSuite) TestFindOrCreate() {
	item, err := suite.typed.Find(test.SimpleItem2.Filter())
	suite.Require().Error(err)
	suite.True(IsNotFound(err))
	suite.Nil(item)
	item, err = suite.typed.FindOrCreate(test.SimpleItem2.Filter(), test.SimpleItem2)
	suite.Require().NoError(err)
	suite.NotNil(item)
	item, err = suite.typed.Find(test.SimpleItem2.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	item2, err := suite.typed.FindOrCreate(test.SimpleItem2.Filter(), test.SimpleItem2)
	suite.Require().NoError(err)
	suite.NotNil(item2)
	suite.Equal(item, item2)
}

func (suite *typedTestSuite) TestCreateFindDelete() {
	err := suite.typed.Create(test.SimpleItem2)
	suite.Require().NoError(err)
	item, err := suite.typed.Find(test.SimpleItem2.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	cacheKey := test.SimpleItem2.CacheKey()
	suite.NotEmpty(cacheKey)
	err = suite.typed.Delete(test.SimpleItem2.Filter(), false)
	suite.Require().NoError(err)
	noItem, err := suite.typed.Find(test.SimpleItem2.Filter())
	suite.Require().Error(err)
	suite.True(IsNotFound(err))
	suite.Nil(noItem)
	err = suite.typed.Delete(test.SimpleItem2.Filter(), false)
	suite.Require().Error(err)
	err = suite.typed.Delete(test.SimpleItem2.Filter(), true)
	suite.Require().NoError(err)
}

func (suite *typedTestSuite) TestCountDeleteAll() {
	suite.Require().NoError(suite.typed.Create(test.SimpleItem1))
	suite.Require().NoError(suite.typed.Create(test.SimpleItem2))
	suite.Require().NoError(suite.typed.Create(test.SimpleItem3))
	count, err := suite.typed.Count(bson.D{})
	suite.NoError(err)
	suite.Equal(int64(3), count)
	suite.NoError(suite.typed.DeleteAll())
	count, err = suite.typed.Count(bson.D{})
	suite.NoError(err)
	suite.Equal(int64(0), count)
}

func (suite *typedTestSuite) TestIterate() {
	suite.Require().NoError(suite.typed.Create(test.SimpleItem1))
	suite.Require().NoError(suite.typed.Create(test.SimpleItem2))
	suite.Require().NoError(suite.typed.Create(test.SimpleItem3))
	count := 0
	var alpha []string
	suite.NoError(suite.typed.Iterate(bson.D{},
		func(item *test.SimpleItem) error {
			alpha = append(alpha, item.Alpha)
			count++
			return nil
		}))
	suite.Equal(3, count)
	suite.Equal([]string{"one", "two", "three"}, alpha)
}

func (suite *typedTestSuite) TestIterateFiltered() {
	suite.Require().NoError(suite.typed.Create(test.SimpleItem1))
	suite.Require().NoError(suite.typed.Create(test.SimpleItem2))
	suite.Require().NoError(suite.typed.Create(test.SimpleItem3))
	count := 0
	var alpha []string
	suite.NoError(suite.typed.Iterate(bson.D{bson.E{Key: "bravo", Value: 2}},
		func(item *test.SimpleItem) error {
			alpha = append(alpha, item.Alpha)
			count++
			return nil
		}))
	suite.Equal(1, count)
	suite.Equal([]string{"two"}, alpha)
}

func (suite *typedTestSuite) TestStringValuesFor() {
	collection, err := suite.access.Collection(context.TODO(), "mdb-typed-collection-string-values", "")
	suite.Require().NoError(err)
	typed := NewTypedCollection[test.SimpleItem](collection)
	suite.NotNil(typed)
	for i := 0; i < 5; i++ {
		suite.Require().NoError(typed.Create(&test.SimpleItem{
			SimpleKey: test.SimpleKey{
				Alpha: fmt.Sprintf("Alpha #%d", i),
				Bravo: i,
			},
			Charlie: "There can be only one",
		}))
	}
	values, err := typed.StringValuesFor("alpha", nil)
	suite.Require().NoError(err)
	suite.Len(values, 5)
	values, err = typed.StringValuesFor("charlie", nil)
	suite.Require().NoError(err)
	suite.Len(values, 1)
	values, err = typed.StringValuesFor("goober", nil)
	suite.Require().NoError(err)
	suite.Len(values, 0)
}

////////////////////////////////////////////////////////////////////////////////

func (suite *typedTestSuite) TestCreateFindDeleteWrapped() {
	collection, err := suite.access.Collection(context.TODO(), "mdb-cached-collection-wrapped-items", "")
	suite.Require().NoError(err)
	wrapped := NewTypedCollection[WrappedItems](collection)
	suite.Require().NotNil(wrapped)
	suite.Require().NoError(wrapped.DeleteAll())
	wrappedItems := MakeWrappedItems()
	suite.Require().NoError(wrapped.Create(wrappedItems))
	foundWrapped, err := wrapped.Find(wrappedItems.Filter())
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
	err = wrapped.Delete(wrappedItems.Filter(), false)
	suite.Require().NoError(err)
	noItem, err := wrapped.Find(wrappedItems.Filter())
	suite.Require().Error(err)
	suite.True(IsNotFound(err))
	suite.Nil(noItem)
	err = wrapped.Delete(wrappedItems.Filter(), false)
	suite.Require().Error(err)
	err = wrapped.Delete(wrappedItems.Filter(), true)
	suite.Require().NoError(err)
}
