//go:build database
// +build database

package mdb

import (
	"context"
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
	var err error
	suite.AccessTestSuite.SetupSuite()
	suite.collection, err = suite.access.Collection(context.TODO(), "test-collection", test.SimpleValidatorJSON)
	suite.Require().NoError(err)
	suite.NotNil(suite.collection)
	suite.Require().NoError(suite.access.Index(suite.collection, NewIndexDescription(true, "alpha")))
}

func (suite *collectionTestSuite) TearDownTest() {
	suite.NoError(suite.collection.DeleteAll())
}

func (suite *collectionTestSuite) TestCollectionValidator() {
	collection, err := suite.access.Collection(context.TODO(), "test-collection-validator", test.SimpleValidatorJSON)
	suite.Require().NoError(err)
	suite.NotNil(collection)
}

func (suite *collectionTestSuite) TestCollectionValidatorFinisher() {
	var finished bool
	collection, err := suite.access.Collection(
		context.TODO(), "test-collection-validator-finisher", test.SimpleValidatorJSON,
		func(access *Access, collection *Collection) error {
			access.Info("Running finisher")
			finished = true
			return nil
		})
	suite.Require().NoError(err)
	suite.NotNil(collection)
	suite.True(finished)
}

func (suite *collectionTestSuite) TestCollectionValidatorFinisherError() {
	collection, err := suite.access.Collection(
		context.TODO(), "test-collection-validator-finisher-error", test.SimpleValidatorJSON,
		func(access *Access, collection *Collection) error {
			return errors.New("fail")
		})
	suite.Error(err)
	suite.Nil(collection)
}

func (suite *collectionTestSuite) TestCreateDuplicate() {
	err := suite.collection.Create(test.SimpleItem1)
	suite.Require().NoError(err)
	item, err := suite.collection.Find(test.SimpleItem1.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	err = suite.collection.Create(test.SimpleItem1)
	suite.Require().Error(err)
	suite.Require().True(suite.access.IsDuplicate(err))
}

func (suite *collectionTestSuite) TestFindNone() {
	item, err := suite.collection.Find(test.SimpleKeyOfTheBeast.Filter())
	suite.Require().Error(err)
	suite.True(suite.collection.IsNotFound(err))
	suite.Nil(item)
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
	suite.True(suite.collection.IsNotFound(err))
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
	count, err := suite.collection.Count(bson.D{})
	suite.NoError(err)
	suite.Equal(int64(3), count)
	suite.NoError(suite.collection.DeleteAll())
	count, err = suite.collection.Count(bson.D{})
	suite.NoError(err)
	suite.Equal(int64(0), count)
}

func (suite *collectionTestSuite) TestIterate() {
	suite.Require().NoError(suite.collection.Create(test.SimpleItem1))
	suite.Require().NoError(suite.collection.Create(test.SimpleItem2))
	suite.Require().NoError(suite.collection.Create(test.SimpleItem3))
	count := 0
	alpha := []string{}
	suite.NoError(suite.collection.Iterate(bson.D{}, func(item interface{}) error {
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
	alpha := []string{}
	suite.NoError(suite.collection.Iterate(bson.D{bson.E{Key: "alpha", Value: "one"}}, func(item interface{}) error {
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

func (suite *collectionTestSuite) TestStringValuesFor() {
	collection, err := suite.access.Collection(context.TODO(), "mdb-collection-string-values", "")
	suite.Require().NoError(err)
	suite.NotNil(collection)
	for i := 0; i < 5; i++ {
		_, err := collection.InsertOne(collection.Context(), &test.SimpleItem{
			SimpleKey: test.SimpleKey{
				Alpha: fmt.Sprintf("Alpha #%d", i),
				Bravo: i,
			},
			Charlie: "There can be only one",
		})
		suite.Require().NoError(err)
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
