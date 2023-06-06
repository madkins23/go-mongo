//go:build database

package mdb

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type collectionTestSuite struct {
	AccessTestSuite
	collection *Collection
}

func TestCollectionSuite(t *testing.T) {
	suite.Run(t, new(collectionTestSuite))
}

func (suite *collectionTestSuite) SetupSuite() {
	suite.AccessTestSuite.SetupSuite()
	suite.collection = suite.ConnectCollection(testCollection, NewIndexDescription(true, "alpha"))
}

func (suite *collectionTestSuite) TearDownTest() {
	suite.NoError(suite.collection.DeleteAll())
}

func (suite *collectionTestSuite) TestCollectionValidator() {
	collection, err := ConnectCollection(suite.access, testCollectionValidation)
	suite.Require().NoError(err)
	suite.NotNil(collection)
	suite.Require().NoError(collection.Create(SimpleItem1))
	err = collection.Create(SimplyInvalid)
	suite.Require().Error(err)
	suite.True(IsValidationFailure(err))
}

func (suite *collectionTestSuite) TestCollectionValidatorFinisher() {
	name := "test-collection-validation-finisher"
	var finished bool
	definition := &CollectionDefinition{
		Name:           name,
		ValidationJSON: SimpleValidatorJSON,
		Finishers: []CollectionFinisher{
			func(access *Access, collection *Collection) error {
				access.Info("Running finisher")
				finished = true
				return nil
			},
		},
	}
	collection, err := ConnectCollection(suite.access, definition)
	suite.Require().NoError(err)
	suite.NotNil(collection)
	suite.True(finished)
	suite.True(suite.Access().CollectionExists(name))
	suite.Require().NoError(collection.Drop())
}

func (suite *collectionTestSuite) TestCollectionValidatorFinisherError() {
	name := "test-collection-validation-finisher-error"
	collection, err := ConnectCollection(suite.access, &CollectionDefinition{
		Name:           "test-collection-validation-finisher-error",
		ValidationJSON: SimpleValidatorJSON,
		Finishers: []CollectionFinisher{
			func(access *Access, collection *Collection) error {
				return errors.New("fail")
			},
		},
	})
	suite.Error(err)
	suite.Nil(collection)
	// Make sure the collection was dropped.
	suite.False(suite.Access().CollectionExists(name))
}

func (suite *collectionTestSuite) TestCreateDuplicate() {
	err := suite.collection.Create(SimpleItem1)
	suite.Require().NoError(err)
	item, err := suite.collection.Find(SimpleItem1.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	suite.NotNil(suite.bsonGetID(item))
	err = suite.collection.Create(SimpleItem1)
	suite.Require().Error(err)
	suite.Require().True(IsDuplicate(err))
}

func (suite *collectionTestSuite) TestFindNone() {
	item, err := suite.collection.Find(SimplyInvalid.Filter())
	suite.Require().Error(err)
	suite.True(IsNotFound(err))
	suite.Nil(item)
}

func (suite *collectionTestSuite) TestFindOrCreate() {
	item, err := suite.collection.Find(SimpleItem2.Filter())
	suite.Require().Error(err)
	suite.True(IsNotFound(err))
	suite.Nil(item)
	item, err = suite.collection.FindOrCreate(SimpleItem2.Filter(), SimpleItem2)
	suite.Require().NoError(err)
	suite.NotNil(item)
	suite.NotNil(suite.bsonGetID(item))
	item, err = suite.collection.Find(SimpleItem2.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	item2, err := suite.collection.FindOrCreate(SimpleItem2.Filter(), SimpleItem2)
	suite.Require().NoError(err)
	suite.NotNil(item2)
	suite.Equal(item, item2)
	suite.Equal(suite.bsonGetID(item), suite.bsonGetID(item2))
}

func (suite *collectionTestSuite) TestCreateFindDelete() {
	suite.Require().NoError(suite.collection.Create(SimpleItem2))
	item, err := suite.collection.Find(SimpleItem2.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	suite.NotNil(suite.bsonGetID(item))
	err = suite.collection.Delete(SimpleItem2.Filter(), false)
	suite.Require().NoError(err)
	noItem, err := suite.collection.Find(SimpleItem2.Filter())
	suite.Require().Error(err)
	suite.True(IsNotFound(err))
	suite.Nil(noItem)
	err = suite.collection.Delete(SimpleItem2.Filter(), false)
	suite.Require().Error(err)
	err = suite.collection.Delete(SimpleItem2.Filter(), true)
	suite.Require().NoError(err)
}

func (suite *collectionTestSuite) TestCountDeleteAll() {
	suite.Require().NoError(suite.collection.Create(SimpleItem1))
	suite.Require().NoError(suite.collection.Create(SimpleItem2))
	suite.Require().NoError(suite.collection.Create(SimpleItem3))
	count, err := suite.collection.Count(NoFilter())
	suite.NoError(err)
	suite.Equal(int64(3), count)
	suite.NoError(suite.collection.DeleteAll())
	count, err = suite.collection.Count(NoFilter())
	suite.NoError(err)
	suite.Equal(int64(0), count)
}

func (suite *collectionTestSuite) TestIterate() {
	suite.Require().NoError(suite.collection.Create(SimpleItem1))
	suite.Require().NoError(suite.collection.Create(SimpleItem2))
	suite.Require().NoError(suite.collection.Create(SimpleItem3))
	count := 0
	var alpha []string
	suite.NoError(suite.collection.Iterate(NoFilter(),
		func(item interface{}) error {
			if bd, ok := item.(bson.D); ok {
				m := bd.Map()
				if a, ok := m["alpha"].(string); ok {
					alpha = append(alpha, a)
				}
			}
			count++
			return nil
		}))
	suite.Equal(3, count)
	suite.Equal([]string{"one", "two", "three"}, alpha)
}

func (suite *collectionTestSuite) TestIterateFiltered() {
	suite.Require().NoError(suite.collection.Create(SimpleItem1))
	suite.Require().NoError(suite.collection.Create(SimpleItem2))
	suite.Require().NoError(suite.collection.Create(SimpleItem3))
	count := 0
	var alpha []string
	suite.NoError(suite.collection.Iterate(bson.D{bson.E{Key: "alpha", Value: "one"}},
		func(item interface{}) error {
			if bd, ok := item.(bson.D); ok {
				m := bd.Map()
				if a, ok := m["alpha"].(string); ok {
					alpha = append(alpha, a)
				}
			}
			count++
			return nil
		}))
	suite.Equal(1, count)
	suite.Equal([]string{"one"}, alpha)
}

func (suite *collectionTestSuite) TestReplace() {
	suite.Require().NoError(suite.collection.Create(SimpleItem1))
	item, err := suite.collection.Find(SimpleItem1.Filter())
	suite.Require().NoError(err)
	suite.bsonFieldEquals(item, "alpha", "one")
	suite.NotNil(suite.bsonGetID(item))
	// Replace with new value:
	suite.Require().NoError(suite.collection.Replace(SimpleItem1.Filter(), SimpleItem1x))
	_, err = suite.collection.Find(SimpleItem1.Filter())     // look for old item
	suite.True(IsNotFound(err))                              // gone
	item, err = suite.collection.Find(SimpleItem1x.Filter()) // look for new item
	suite.Require().NoError(err)                             // found
	suite.Require().NotNil(item)
	suite.bsonFieldEquals(item, "alpha", "xRay")
	suite.NotNil(suite.bsonGetID(item))
	// Replace with same value:
	err = suite.collection.Replace(SimpleItem1x.Filter(), SimpleItem1x)
	suite.Require().ErrorIs(err, errNoItemModified)
	item, err = suite.collection.Find(SimpleItem1x.Filter())
	suite.Require().NoError(err)
	suite.bsonFieldEquals(item, "alpha", "xRay")
	// No match for filter:
	item, err = suite.collection.Find(SimpleItem3.Filter())
	suite.True(IsNotFound(err))
	suite.ErrorIs(suite.collection.Replace(SimpleItem3.Filter(), SimpleItem3), errNoItemMatch)
	// Upsert new item:
	suite.NoError(suite.collection.Replace(NoFilter(), SimpleItem3))
	item, err = suite.collection.Find(SimpleItem3.Filter())
	suite.Require().NoError(err)
	suite.bsonFieldEquals(item, "alpha", "three")
	suite.NotNil(suite.bsonGetID(item))
}

func (suite *collectionTestSuite) TestUpdate() {
	suite.Require().NoError(suite.collection.Create(SimpleItem1))
	item, err := suite.collection.Find(SimpleItem1.Filter())
	suite.Require().NoError(err)
	suite.bsonFieldEquals(item, "alpha", "one")
	suite.bsonFieldEquals(item, "charlie", SimpleCharlie1)
	suite.bsonFieldEquals(item, "delta", int32(1))
	// Set charlie and delta fields:
	suite.Require().NoError(
		suite.collection.Update(SimpleItem1.Filter(), bson.M{
			"$set": bson.M{"charlie": "One more time"},
			"$inc": bson.M{"delta": 2},
		}))
	item, err = suite.collection.Find(SimpleItem1.Filter())
	suite.Require().NoError(err)
	suite.Require().NotNil(item)
	suite.bsonFieldEquals(item, "charlie", "One more time")
	suite.bsonFieldEquals(item, "delta", int32(3))
	// No match for filter:
	item, err = suite.collection.Find(SimpleItem3.Filter())
	suite.True(IsNotFound(err))
	suite.ErrorIs(suite.collection.Update(SimpleItem3.Filter(), bson.M{
		"$set": bson.M{"charlie": "Horse"},
		"$inc": bson.M{"delta": 7},
	}), errNoItemMatch)
}

func (suite *collectionTestSuite) TestStringValuesFor() {
	collection, err := ConnectCollection(suite.access, testCollectionStringValues)
	suite.Require().NoError(err)
	suite.NotNil(collection)
	for i := 0; i < 5; i++ {
		suite.Require().NoError(
			collection.Create(&SimpleItem{
				Alpha:   fmt.Sprintf("Alpha #%d", i),
				Bravo:   i,
				Charlie: "There can be only one",
			}))
	}
	values, err := collection.StringValuesFor("alpha", nil)
	suite.Require().NoError(err)
	suite.Len(values, 5)
	values, err = collection.StringValuesFor("charlie", nil)
	suite.Require().NoError(err)
	suite.Len(values, 1)
	values, err = collection.StringValuesFor("goober", nil)
	suite.Require().NoError(err)
	suite.Len(values, 0)
}

////////////////////////////////////////////////////////////////////////////////

func (suite *collectionTestSuite) bsonFieldEquals(item interface{}, field string, value interface{}) {
	suite.Require().NotNil(item)
	itemD, ok := item.(bson.D)
	suite.Require().True(ok)
	fldVal, found := itemD.Map()[field]
	suite.Require().True(found)
	suite.Equal(value, fldVal)
}

func (suite *collectionTestSuite) bsonGetID(item interface{}) primitive.ObjectID {
	suite.Require().NotNil(item)
	itemD, ok := item.(bson.D)
	suite.Require().True(ok)
	fldVal, found := itemD.Map()["_id"]
	suite.Require().True(found)
	suite.NotNil(fldVal)
	objID, ok := fldVal.(primitive.ObjectID)
	suite.Require().True(ok)
	return objID
}
