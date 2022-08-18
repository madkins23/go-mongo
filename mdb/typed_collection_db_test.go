//go:build database

package mdb

import (
	"fmt"
	"testing"

	"github.com/madkins23/go-type/reg"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

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
	var err error
	suite.typed, err = ConnectTypedCollection[test.SimpleItem](suite.access, testCollectionValidation)
	suite.Require().NoError(err)
	suite.NotNil(suite.typed)
	suite.Require().NoError(suite.typed.DeleteAll())
	suite.Require().NoError(suite.access.Index(&suite.typed.Collection, NewIndexDescription(true, "alpha")))
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
	suite.NotNil(item.ID)
	err = suite.typed.Create(test.SimpleItem1)
	suite.Require().Error(err)
	suite.Require().True(IsDuplicate(err))
}

func (suite *typedTestSuite) TestFindNone() {
	item, err := suite.typed.Find(test.SimplyInvalid.Filter())
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
	suite.NotNil(item.ID)
	itemID := item.ID
	item, err = suite.typed.Find(test.SimpleItem2.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	suite.NotNil(item.ID)
	suite.Equal(itemID, item.ID)
	item2, err := suite.typed.FindOrCreate(test.SimpleItem2.Filter(), test.SimpleItem2)
	suite.Require().NoError(err)
	suite.NotNil(item2)
	suite.Equal(item, item2)
	suite.Equal(itemID, item2.ID)
}

func (suite *typedTestSuite) TestCreateFindDelete() {
	err := suite.typed.Create(test.SimpleItem2)
	suite.Require().NoError(err)
	item, err := suite.typed.Find(test.SimpleItem2.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	suite.NotNil(item.ID)
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
	count, err := suite.typed.Count(NoFilter())
	suite.NoError(err)
	suite.Equal(int64(3), count)
	suite.NoError(suite.typed.DeleteAll())
	count, err = suite.typed.Count(NoFilter())
	suite.NoError(err)
	suite.Equal(int64(0), count)
}

func (suite *typedTestSuite) TestIterate() {
	suite.Require().NoError(suite.typed.Create(test.SimpleItem1))
	suite.Require().NoError(suite.typed.Create(test.SimpleItem2))
	suite.Require().NoError(suite.typed.Create(test.SimpleItem3))
	count := 0
	var alpha []string
	suite.NoError(suite.typed.Iterate(NoFilter(),
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

func (suite *typedTestSuite) TestReplace() {
	suite.Require().NoError(suite.typed.Create(test.SimpleItem1))
	item, err := suite.typed.Find(test.SimpleItem1.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	suite.NotNil(item.ID)
	itemID := item.ID
	suite.Equal("one", item.Alpha)
	// Replace with new value:
	suite.Require().NoError(suite.typed.Replace(test.SimpleItem1, test.SimpleItem1x))
	_, err = suite.typed.Find(test.SimpleItem1.Filter())     // look for old item
	suite.True(IsNotFound(err))                              // gone
	item, err = suite.typed.Find(test.SimpleItem1x.Filter()) // look for new item
	suite.Require().NoError(err)                             // found
	suite.Require().NotNil(item)
	suite.NotNil(item.ID)
	suite.Equal(itemID, item.ID)
	suite.Equal("xRay", item.Alpha)
	// Replace with same value:
	err = suite.typed.Replace(test.SimpleItem1x, test.SimpleItem1x)
	suite.Require().ErrorIs(err, errNoItemModified)
	item, err = suite.typed.Find(test.SimpleItem1x.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	suite.NotNil(item.ID)
	suite.Equal(itemID, item.ID)
	suite.Equal("xRay", item.Alpha)
	// No match for filter:
	item, err = suite.typed.Find(test.SimpleItem3.Filter())
	suite.True(IsNotFound(err))
	suite.ErrorIs(suite.typed.Replace(test.SimpleItem3, test.SimpleItem3), errNoItemMatch)
	// Upsert new item:
	suite.NoError(suite.typed.Replace(NoFilter(), test.SimpleItem3))
	item, err = suite.typed.Find(test.SimpleItem3.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	suite.NotNil(item.ID)
	suite.Equal("three", item.Alpha)
	suite.Equal(itemID, item.ID)
}

func (suite *typedTestSuite) TestUpdate() {
	suite.Require().NoError(suite.typed.Create(test.SimpleItem1))
	item, err := suite.typed.Find(test.SimpleItem1.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	suite.NotNil(item.ID)
	itemID := item.ID
	suite.Equal("one", item.Alpha)
	suite.Equal(test.SimpleCharlie1, item.Charlie)
	suite.Equal(1, item.Delta)
	// Set charlie and delta fields:
	suite.Require().NoError(
		suite.typed.Update(test.SimpleItem1.Filter(), bson.M{
			"$set": bson.M{"charlie": "One more time"},
			"$inc": bson.M{"delta": 2},
		}))
	item, err = suite.typed.Find(test.SimpleItem1.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	suite.Equal(itemID, item.ID)
	suite.Equal("One more time", item.Charlie)
	suite.Equal(3, item.Delta)
	// No match for filter:
	item, err = suite.typed.Find(test.SimpleItem3.Filter())
	suite.True(IsNotFound(err))
	suite.ErrorIs(suite.typed.Update(test.SimpleItem3.Filter(), bson.M{
		"$set": bson.M{"charlie": "Horse"},
		"$inc": bson.M{"delta": 7},
	}), errNoItemMatch)
}

func (suite *typedTestSuite) TestStringValuesFor() {
	typed, err := ConnectTypedCollection[test.SimpleItem](suite.access, testCollectionStringValues)
	suite.Require().NoError(err)
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
	wrapped, err := ConnectTypedCollection[WrappedItems](suite.access, testCollectionWrapped)
	suite.Require().NoError(err)
	suite.Require().NotNil(wrapped)
	suite.Require().NoError(wrapped.DeleteAll())
	wrappedItems := MakeWrappedItems()
	suite.Require().NoError(wrapped.Create(wrappedItems))
	foundWrapped, err := wrapped.Find(wrappedItems.Filter())
	suite.Require().NoError(err)
	suite.Require().NotNil(foundWrapped)
	suite.NotNil(foundWrapped.ID)
	// Zero out the object ID before testing equality.
	foundWrapped.OID = primitive.ObjectID{}
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
