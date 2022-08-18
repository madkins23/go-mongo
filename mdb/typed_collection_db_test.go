//go:build database

package mdb

import (
	"fmt"
	"testing"

	"github.com/madkins23/go-type/reg"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type typedTestSuite struct {
	AccessTestSuite
	typed *TypedCollection[SimpleItem]
}

func TestTypedSuite(t *testing.T) {
	suite.Run(t, new(typedTestSuite))
}

func (suite *typedTestSuite) SetupSuite() {
	suite.AccessTestSuite.SetupSuite()
	reg.Highlander().Clear()
	suite.Require().NoError(RegisterWrapped())
	var err error
	suite.typed, err = ConnectTypedCollection[SimpleItem](suite.access, testCollectionValidation)
	suite.Require().NoError(err)
	suite.NotNil(suite.typed)
	suite.Require().NoError(suite.typed.DeleteAll())
	suite.Require().NoError(suite.access.Index(&suite.typed.Collection, NewIndexDescription(true, "alpha")))
}

func (suite *typedTestSuite) TearDownTest() {
	suite.NoError(suite.typed.DeleteAll())
}

func (suite *typedTestSuite) TestCreateDuplicate() {
	err := suite.typed.Create(SimpleItem1)
	suite.Require().NoError(err)
	item, err := suite.typed.Find(SimpleItem1.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	suite.NotNil(item.ID)
	err = suite.typed.Create(SimpleItem1)
	suite.Require().Error(err)
	suite.Require().True(IsDuplicate(err))
}

func (suite *typedTestSuite) TestFindNone() {
	item, err := suite.typed.Find(SimplyInvalid.Filter())
	suite.Require().Error(err)
	suite.True(IsNotFound(err))
	suite.Nil(item)
}

func (suite *typedTestSuite) TestFindOrCreate() {
	item, err := suite.typed.Find(SimpleItem2.Filter())
	suite.Require().Error(err)
	suite.True(IsNotFound(err))
	suite.Nil(item)
	item, err = suite.typed.FindOrCreate(SimpleItem2.Filter(), SimpleItem2)
	suite.Require().NoError(err)
	suite.NotNil(item)
	suite.NotNil(item.ID())
	itemID := item.ID()
	item, err = suite.typed.Find(SimpleItem2.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	suite.NotNil(item.ID())
	suite.Equal(itemID, item.ID())
	item2, err := suite.typed.FindOrCreate(SimpleItem2.Filter(), SimpleItem2)
	suite.Require().NoError(err)
	suite.NotNil(item2)
	suite.Equal(item, item2)
	suite.Equal(itemID, item2.ObjectID)
}

func (suite *typedTestSuite) TestCreateFindDelete() {
	err := suite.typed.Create(SimpleItem2)
	suite.Require().NoError(err)
	item, err := suite.typed.Find(SimpleItem2.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	suite.NotNil(item.ID)
	err = suite.typed.Delete(SimpleItem2.Filter(), false)
	suite.Require().NoError(err)
	noItem, err := suite.typed.Find(SimpleItem2.Filter())
	suite.Require().Error(err)
	suite.True(IsNotFound(err))
	suite.Nil(noItem)
	err = suite.typed.Delete(SimpleItem2.Filter(), false)
	suite.Require().Error(err)
	err = suite.typed.Delete(SimpleItem2.Filter(), true)
	suite.Require().NoError(err)
}

func (suite *typedTestSuite) TestCountDeleteAll() {
	suite.Require().NoError(suite.typed.Create(SimpleItem1))
	suite.Require().NoError(suite.typed.Create(SimpleItem2))
	suite.Require().NoError(suite.typed.Create(SimpleItem3))
	count, err := suite.typed.Count(NoFilter())
	suite.NoError(err)
	suite.Equal(int64(3), count)
	suite.NoError(suite.typed.DeleteAll())
	count, err = suite.typed.Count(NoFilter())
	suite.NoError(err)
	suite.Equal(int64(0), count)
}

func (suite *typedTestSuite) TestIterate() {
	suite.Require().NoError(suite.typed.Create(SimpleItem1))
	suite.Require().NoError(suite.typed.Create(SimpleItem2))
	suite.Require().NoError(suite.typed.Create(SimpleItem3))
	count := 0
	var alpha []string
	suite.NoError(suite.typed.Iterate(NoFilter(),
		func(item *SimpleItem) error {
			alpha = append(alpha, item.Alpha)
			count++
			return nil
		}))
	suite.Equal(3, count)
	suite.Equal([]string{"one", "two", "three"}, alpha)
}

func (suite *typedTestSuite) TestIterateFiltered() {
	suite.Require().NoError(suite.typed.Create(SimpleItem1))
	suite.Require().NoError(suite.typed.Create(SimpleItem2))
	suite.Require().NoError(suite.typed.Create(SimpleItem3))
	count := 0
	var alpha []string
	suite.NoError(suite.typed.Iterate(bson.D{bson.E{Key: "bravo", Value: 2}},
		func(item *SimpleItem) error {
			alpha = append(alpha, item.Alpha)
			count++
			return nil
		}))
	suite.Equal(1, count)
	suite.Equal([]string{"two"}, alpha)
}

func (suite *typedTestSuite) TestReplace() {
	suite.Require().NoError(suite.typed.Create(SimpleItem1))
	item, err := suite.typed.Find(SimpleItem1.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	suite.NotNil(item.ID())
	itemID := item.ID()
	suite.Equal("one", item.Alpha)
	// Replace with new value:
	suite.Require().NoError(suite.typed.Replace(SimpleItem1, SimpleItem1x))
	_, err = suite.typed.Find(SimpleItem1.Filter())     // look for old item
	suite.True(IsNotFound(err))                         // gone
	item, err = suite.typed.Find(SimpleItem1x.Filter()) // look for new item
	suite.Require().NoError(err)                        // found
	suite.Require().NotNil(item)
	suite.NotNil(item.ID())
	suite.Equal(itemID, item.ID())
	suite.Equal("xRay", item.Alpha)
	// Replace with same value:
	err = suite.typed.Replace(SimpleItem1x, SimpleItem1x)
	suite.Require().ErrorIs(err, errNoItemModified)
	item, err = suite.typed.Find(SimpleItem1x.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	suite.NotNil(item.ID())
	suite.Equal(itemID, item.ID())
	suite.Equal("xRay", item.Alpha)
	// No match for filter:
	item, err = suite.typed.Find(SimpleItem3.Filter())
	suite.True(IsNotFound(err))
	suite.ErrorIs(suite.typed.Replace(SimpleItem3, SimpleItem3), errNoItemMatch)
	// Upsert new item:
	suite.NoError(suite.typed.Replace(NoFilter(), SimpleItem3))
	item, err = suite.typed.Find(SimpleItem3.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	suite.NotNil(item.ID)
	suite.Equal("three", item.Alpha)
	suite.Equal(itemID, item.ID())
}

func (suite *typedTestSuite) TestStringValuesFor() {
	typed, err := ConnectTypedCollection[SimpleItem](suite.access, testCollectionStringValues)
	suite.Require().NoError(err)
	suite.NotNil(typed)
	for i := 0; i < 5; i++ {
		suite.Require().NoError(typed.Create(&SimpleItem{
			Alpha:   fmt.Sprintf("Alpha #%d", i),
			Bravo:   i,
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

func (suite *typedTestSuite) TestUpdate() {
	suite.Require().NoError(suite.typed.Create(SimpleItem1))
	item, err := suite.typed.Find(SimpleItem1.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	suite.NotNil(item.ID())
	itemID := item.ID()
	suite.Equal("one", item.Alpha)
	suite.Equal(SimpleCharlie1, item.Charlie)
	suite.Equal(1, item.Delta)
	// Set charlie and delta fields:
	suite.Require().NoError(
		suite.typed.Update(SimpleItem1.Filter(), bson.M{
			"$set": bson.M{"charlie": "One more time"},
			"$inc": bson.M{"delta": 2},
		}))
	item, err = suite.typed.Find(SimpleItem1.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	suite.Equal(itemID, item.ID())
	suite.Equal("One more time", item.Charlie)
	suite.Equal(3, item.Delta)
	// No match for filter:
	item, err = suite.typed.Find(SimpleItem3.Filter())
	suite.True(IsNotFound(err))
	suite.ErrorIs(suite.typed.Update(SimpleItem3.Filter(), bson.M{
		"$set": bson.M{"charlie": "Horse"},
		"$inc": bson.M{"delta": 7},
	}), errNoItemMatch)
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
	foundWrapped.ObjectID = primitive.ObjectID{}
	suite.Equal(wrappedItems, foundWrapped)
	suite.Equal(ValueText, foundWrapped.Single.Get().String())
	for _, item := range foundWrapped.Array {
		switch item.Get().Key() {
		case "text":
			suite.Equal(ValueText, item.Get().String())
		case "numeric":
			if numVal, ok := item.Get().(*NumericValue); ok {
				suite.Equal(ValueNumber, numVal.Number)
			} else {
				suite.Fail("Not NumericValue: " + item.Get().String())
			}
		case "random":
			random := item.Get().String()
			fmt.Printf("Random:  %s\n", random)
			suite.True(len(random) >= RandomMinimum)
			suite.True(len(random) <= RandomMaximum)
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
