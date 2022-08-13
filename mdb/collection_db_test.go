//go:build database
// +build database

package mdb

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/madkins23/go-mongo/test"
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
	var err error
	suite.collection, err = ConnectCollection(suite.access, testCollection)
	suite.Require().NoError(err)
	suite.NotNil(suite.collection)
	suite.Require().NoError(suite.collection.DeleteAll())
	suite.Require().NoError(suite.access.Index(suite.collection, NewIndexDescription(true, "alpha")))
}

func (suite *collectionTestSuite) TearDownTest() {
	suite.NoError(suite.collection.DeleteAll())
}

func (suite *collectionTestSuite) TestCollectionValidator() {
	collection, err := ConnectCollection(suite.access, testCollectionValidation)
	suite.Require().NoError(err)
	suite.NotNil(collection)
	suite.Require().NoError(collection.Create(test.SimpleItem1))
	err = collection.Create(test.SimplyInvalid)
	suite.Require().Error(err)
	suite.True(IsValidationFailure(err))
}

func (suite *collectionTestSuite) TestCollectionValidatorFinisher() {
	name := "test-collection-validation-finisher"
	var finished bool
	definition := &CollectionDefinition{
		name:           name,
		validationJSON: test.SimpleValidatorJSON,
		finishers: []CollectionFinisher{
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
		name:           "test-collection-validation-finisher-error",
		validationJSON: test.SimpleValidatorJSON,
		finishers: []CollectionFinisher{
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
	err := suite.collection.Create(test.SimpleItem1)
	suite.Require().NoError(err)
	item, err := suite.collection.Find(test.SimpleItem1.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	err = suite.collection.Create(test.SimpleItem1)
	suite.Require().Error(err)
	suite.Require().True(IsDuplicate(err))
}

func (suite *collectionTestSuite) TestFindNone() {
	item, err := suite.collection.Find(test.SimplyInvalid.Filter())
	suite.Require().Error(err)
	suite.True(IsNotFound(err))
	suite.Nil(item)
}

func (suite *collectionTestSuite) TestFindOrCreate() {
	item, err := suite.collection.Find(test.SimpleItem2.Filter())
	suite.Require().Error(err)
	suite.True(IsNotFound(err))
	suite.Nil(item)
	item, err = suite.collection.FindOrCreate(test.SimpleItem2.Filter(), test.SimpleItem2)
	suite.Require().NoError(err)
	suite.NotNil(item)
	item, err = suite.collection.Find(test.SimpleItem2.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	item2, err := suite.collection.FindOrCreate(test.SimpleItem2.Filter(), test.SimpleItem2)
	suite.Require().NoError(err)
	suite.NotNil(item2)
	suite.Equal(item, item2)
}

func (suite *collectionTestSuite) TestCreateFindDelete() {
	suite.Require().NoError(suite.collection.Create(test.SimpleItem2))
	item, err := suite.collection.Find(test.SimpleItem2.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	cacheKey := test.SimpleItem2.CacheKey()
	suite.NotEmpty(cacheKey)
	err = suite.collection.Delete(test.SimpleItem2.Filter(), false)
	suite.Require().NoError(err)
	noItem, err := suite.collection.Find(test.SimpleItem2.Filter())
	suite.Require().Error(err)
	suite.True(IsNotFound(err))
	suite.Nil(noItem)
	err = suite.collection.Delete(test.SimpleItem2.Filter(), false)
	suite.Require().Error(err)
	err = suite.collection.Delete(test.SimpleItem2.Filter(), true)
	suite.Require().NoError(err)
}

func (suite *collectionTestSuite) TestCountDeleteAll() {
	suite.Require().NoError(suite.collection.Create(test.SimpleItem1))
	suite.Require().NoError(suite.collection.Create(test.SimpleItem2))
	suite.Require().NoError(suite.collection.Create(test.SimpleItem3))
	count, err := suite.collection.Count(NoFilter())
	suite.NoError(err)
	suite.Equal(int64(3), count)
	suite.NoError(suite.collection.DeleteAll())
	count, err = suite.collection.Count(NoFilter())
	suite.NoError(err)
	suite.Equal(int64(0), count)
}

func (suite *collectionTestSuite) TestIterate() {
	suite.Require().NoError(suite.collection.Create(test.SimpleItem1))
	suite.Require().NoError(suite.collection.Create(test.SimpleItem2))
	suite.Require().NoError(suite.collection.Create(test.SimpleItem3))
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
	suite.Require().NoError(suite.collection.Create(test.SimpleItem1))
	suite.Require().NoError(suite.collection.Create(test.SimpleItem2))
	suite.Require().NoError(suite.collection.Create(test.SimpleItem3))
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
	suite.Require().NoError(suite.collection.Create(test.SimpleItem1))
	item, err := suite.collection.Find(test.SimpleItem1.Filter())
	suite.Require().NoError(err)
	suite.checkBsonField(item, "alpha", "one")
	// Replace with new value:
	suite.Require().NoError(suite.collection.Replace(test.SimpleItem1.Filter(), test.SimpleItem1x))
	_, err = suite.collection.Find(test.SimpleItem1.Filter())     // look for old item
	suite.True(IsNotFound(err))                                   // gone
	item, err = suite.collection.Find(test.SimpleItem1x.Filter()) // look for new item
	suite.Require().NoError(err)                                  // found
	suite.Require().NotNil(item)
	suite.checkBsonField(item, "alpha", "xRay")
	// Replace with same value:
	err = suite.collection.Replace(test.SimpleItem1x.Filter(), test.SimpleItem1x)
	suite.Require().ErrorIs(err, errNoItemModified)
	item, err = suite.collection.Find(test.SimpleItem1x.Filter())
	suite.Require().NoError(err)
	suite.checkBsonField(item, "alpha", "xRay")
	// No match for filter:
	item, err = suite.collection.Find(test.SimpleItem3.Filter())
	suite.True(IsNotFound(err))
	suite.ErrorIs(suite.collection.Replace(test.SimpleItem3.Filter(), test.SimpleItem3), errNoItemMatch)
	// Upsert new item:
	suite.NoError(suite.collection.Replace(NoFilter(), test.SimpleItem3))
	item, err = suite.collection.Find(test.SimpleItem3.Filter())
	suite.Require().NoError(err)
	suite.checkBsonField(item, "alpha", "three")
}

func (suite *collectionTestSuite) TestUpdate() {
	suite.Require().NoError(suite.collection.Create(test.SimpleItem1))
	item, err := suite.collection.Find(test.SimpleItem1.Filter())
	suite.Require().NoError(err)
	suite.checkBsonField(item, "alpha", "one")
	suite.checkBsonField(item, "charlie", test.SimpleCharlie1)
	suite.checkBsonField(item, "delta", int32(1))
	// Set charlie and delta fields:
	suite.Require().NoError(
		suite.collection.Update(test.SimpleItem1.Filter(), bson.M{
			"$set": bson.M{"charlie": "One more time"},
			"$inc": bson.M{"delta": 2},
		}))
	item, err = suite.collection.Find(test.SimpleItem1.Filter())
	suite.Require().NoError(err)
	suite.Require().NotNil(item)
	suite.checkBsonField(item, "charlie", "One more time")
	suite.checkBsonField(item, "delta", int32(3))
	// No match for filter:
	item, err = suite.collection.Find(test.SimpleItem3.Filter())
	suite.True(IsNotFound(err))
	suite.ErrorIs(suite.collection.Update(test.SimpleItem3.Filter(), bson.M{
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
			collection.Create(&test.SimpleItem{
				SimpleKey: test.SimpleKey{
					Alpha: fmt.Sprintf("Alpha #%d", i),
					Bravo: i,
				},
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

func (suite *collectionTestSuite) checkBsonField(item interface{}, field string, value interface{}) {
	suite.Require().NotNil(item)
	itemD, ok := item.(bson.D)
	suite.Require().True(ok)
	fldVal, found := itemD.Map()[field]
	suite.Require().True(found)
	suite.Equal(fldVal, value)
}
