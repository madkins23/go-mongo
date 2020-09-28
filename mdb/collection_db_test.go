// +build database

package mdb

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type collectionTestSuite struct {
	AccessTestSuite
}

func TestCollectionSuite(t *testing.T) {
	suite.Run(t, new(collectionTestSuite))
}

func (suite *collectionTestSuite) TestCollection() {
	collection, err := suite.access.Collection("mdb-collection", "")
	suite.Require().NoError(err)
	suite.NotNil(collection)
}

func (suite *collectionTestSuite) TestCollectionValidator() {
	collection, err := suite.access.Collection("mdb-collection", testValidatorJSON)
	suite.Require().NoError(err)
	suite.NotNil(collection)
}

func (suite *collectionTestSuite) TestCollectionValidatorFinisher() {
	var finished bool
	collection, err := suite.access.Collection(
		"mdb-collection-finisher", testValidatorJSON,
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
		"mdb-collection-finisher-error", testValidatorJSON,
		func(access *Access, collection *Collection) error {
			return errors.New("fail")
		})
	suite.Error(err)
	suite.Nil(collection)
}

func (suite *collectionTestSuite) TestStringValuesFor() {
	collection, err := suite.access.Collection("mdb-collection-string-values", "")
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
