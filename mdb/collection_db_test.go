// +build database

package mdb

import (
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"

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
		func(access *Access, collection *mongo.Collection) error {
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
		func(access *Access, collection *mongo.Collection) error {
			return errors.New("fail")
		})
	suite.Error(err)
	suite.Nil(collection)
}

var testValidatorJSON = `{
	"$jsonSchema": {
		"bsonType": "object",
		"required": ["alpha", "bravo"],
		"properties": {
			"alpha": {
				"bsonType": "string"
			},
			"bravo": {
				"bsonType": "int"
			}
		}
	}
}`
