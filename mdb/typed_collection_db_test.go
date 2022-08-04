//go:build database

package mdb

import (
	"context"
	"fmt"
	"testing"

	"github.com/madkins23/go-type/reg"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/madkins23/go-mongo/mdbson"
	"github.com/madkins23/go-mongo/test"
)

type typedTestSuite struct {
	AccessTestSuite
	typed   *TypedCollection[test.SimpleItem]
	wrapped *TypedCollection[WrappedItems]
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
	suite.wrapped = NewTypedCollection[WrappedItems](collection)
	suite.Require().NotNil(suite.wrapped)
	suite.Require().NoError(suite.wrapped.DeleteAll())
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
	suite.Require().True(suite.access.IsDuplicate(err))
}

func (suite *typedTestSuite) TestFindNone() {
	item, err := suite.typed.Find(test.SimpleKeyOfTheBeast.Filter())
	suite.Require().Error(err)
	suite.True(suite.typed.IsNotFound(err))
	suite.Nil(item)
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
	suite.True(suite.typed.IsNotFound(err))
	suite.Nil(noItem)
	err = suite.typed.Delete(test.SimpleItem2.Filter(), false)
	suite.Require().Error(err)
	err = suite.typed.Delete(test.SimpleItem2.Filter(), true)
	suite.Require().NoError(err)
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

func (suite *typedTestSuite) TestCreateFindDeleteWrapped() {
	wrapped := MakeWrappedItems()
	suite.Require().NoError(suite.typed.Create(wrapped))
	foundWrapped, err := suite.wrapped.Find(wrapped.Filter())
	suite.Require().NoError(err)
	suite.Require().NotNil(foundWrapped)
	suite.Assert().Equal(wrapped, foundWrapped)
	suite.Assert().Equal(test.ValueText, foundWrapped.Single.Get().String())
	for _, item := range foundWrapped.Array {
		switch item.Get().Key() {
		case "text":
			suite.Assert().Equal(test.ValueText, item.Get().String())
		case "numeric":
			if numVal, ok := item.Get().(*test.NumericValue); ok {
				suite.Assert().Equal(test.ValueNumber, numVal.Number)
			} else {
				suite.Assert().Fail("Not NumericValue: " + item.Get().String())
			}
		case "random":
			random := item.Get().String()
			fmt.Printf("Random:  %s\n", random)
			suite.Assert().True(len(random) >= test.RandomMinimum)
			suite.Assert().True(len(random) <= test.RandomMaximum)
		default:
			suite.Assert().Fail("Unknown item key: '" + item.Get().Key() + "'")
		}
	}
	for key, item := range foundWrapped.Map {
		suite.Assert().Equal(key, item.Get().Key())
	}
	cacheKey := wrapped.CacheKey()
	suite.NotEmpty(cacheKey)
	err = suite.typed.Delete(wrapped.Filter(), false)
	suite.Require().NoError(err)
	noItem, err := suite.typed.Find(wrapped.Filter())
	suite.Require().Error(err)
	suite.True(suite.typed.IsNotFound(err))
	suite.Nil(noItem)
	err = suite.typed.Delete(wrapped.Filter(), false)
	suite.Require().Error(err)
	err = suite.typed.Delete(wrapped.Filter(), true)
	suite.Require().NoError(err)
}

type WrappedItems struct {
	test.SimpleItem `bson:"inline"`
	Single          *mdbson.Wrapper[test.Wrappable]
	Array           []*mdbson.Wrapper[test.Wrappable]
	Map             map[string]*mdbson.Wrapper[test.Wrappable]
}

func (wi *WrappedItems) Filter() bson.D {
	return bson.D{
		{"alpha", wi.Alpha},
		{"bravo", wi.Bravo},
	}
}

////////////////////////////////////////////////////////////////////////////////

func MakeWrappedItems() *WrappedItems {
	items := []*mdbson.Wrapper[test.Wrappable]{
		mdbson.Wrap[test.Wrappable](&test.TextValue{Text: test.ValueText}),
		mdbson.Wrap[test.Wrappable](&test.NumericValue{Number: test.ValueNumber}),
		mdbson.Wrap[test.Wrappable](&test.RandomValue{}),
	}
	wrapped := &WrappedItems{
		SimpleItem: test.SimpleItem{
			SimpleKey: test.SimpleKey{
				Alpha: "Wrapped",
				Bravo: 23,
			},
			Charlie: "Need this to pass validation",
		},
		Single: items[0],
		Array:  items,
		Map:    make(map[string]*mdbson.Wrapper[test.Wrappable], len(items)),
	}
	for _, item := range items {
		wrapped.Map[item.Get().Key()] = item
	}
	return wrapped
}
