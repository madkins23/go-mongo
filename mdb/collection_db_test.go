// +build database

package mdb

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
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
	suite.collection, err = suite.access.Collection(context.TODO(), "test-collection", testValidatorJSON)
	suite.Require().NoError(err)
	suite.NotNil(suite.collection)
	suite.Require().NoError(suite.access.Index(suite.collection, NewIndexDescription(true, "alpha")))
}

func (suite *collectionTestSuite) TestCollection() {
	collection, err := suite.access.Collection(context.TODO(), "mdb-collection", "")
	suite.Require().NoError(err)
	suite.NotNil(collection)
}

func (suite *collectionTestSuite) TestCollectionValidator() {
	collection, err := suite.access.Collection(context.TODO(), "mdb-collection", testValidatorJSON)
	suite.Require().NoError(err)
	suite.NotNil(collection)
}

func (suite *collectionTestSuite) TestCollectionValidatorFinisher() {
	var finished bool
	collection, err := suite.access.Collection(
		context.TODO(), "mdb-collection-finisher", testValidatorJSON,
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
		context.TODO(), "mdb-collection-finisher-error", testValidatorJSON,
		func(access *Access, collection *Collection) error {
			return errors.New("fail")
		})
	suite.Error(err)
	suite.Nil(collection)
}

func (suite *collectionTestSuite) TestCreateDuplicate() {
	tk := &TestKey{
		Alpha: "two",
		Bravo: 2,
	}
	ti := &testItem{
		TestKey: *tk,
		Charlie: "Two is too much",
	}
	err := suite.collection.Create(ti)
	suite.Require().NoError(err)
	item, err := suite.collection.Find(tk.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	err = suite.collection.Create(ti)
	suite.Require().Error(err)
	suite.Require().True(suite.access.Duplicate(err))
}

func (suite *collectionTestSuite) TestFindNone() {
	tk := &TestKey{
		Alpha: "beast",
		Bravo: 666,
	}
	item, err := suite.collection.Find(tk.Filter())
	suite.Require().Error(err)
	suite.True(suite.collection.NotFound(err))
	suite.Nil(item)
}

func (suite *collectionTestSuite) TestCreateFindDelete() {
	tk := &TestKey{
		Alpha: "one",
		Bravo: 1,
	}
	ti := &testItem{
		TestKey: *tk,
		Charlie: "One is the loneliest number",
	}
	err := suite.collection.Create(ti)
	suite.Require().NoError(err)
	item, err := suite.collection.Find(tk.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	cacheKey := tk.CacheKey()
	suite.NotEmpty(cacheKey)
	err = suite.collection.Delete(tk.Filter(), false)
	suite.Require().NoError(err)
	noItem, err := suite.collection.Find(tk.Filter())
	suite.Require().Error(err)
	suite.True(suite.collection.NotFound(err))
	suite.Nil(noItem)
	err = suite.collection.Delete(tk.Filter(), false)
	suite.Require().Error(err)
	err = suite.collection.Delete(tk.Filter(), true)
	suite.Require().NoError(err)
}

func (suite *collectionTestSuite) TestStringValuesFor() {
	collection, err := suite.access.Collection(context.TODO(), "mdb-collection-string-values", "")
	suite.Require().NoError(err)
	suite.NotNil(collection)
	for i := 0; i < 5; i++ {
		_, err := collection.InsertOne(collection.Context(), &testItem{
			TestKey: TestKey{
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
