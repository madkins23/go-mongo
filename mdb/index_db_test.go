// +build database

package mdb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type indexTestSuite struct {
	AccessTestSuite
	collection *Collection
}

func TestIndexSuite(t *testing.T) {
	suite.Run(t, new(indexTestSuite))
}

func (suite *indexTestSuite) SetupTest() {
	var err error
	suite.collection, err = suite.access.Collection(context.TODO(), "test-index-collection", testValidatorJSON)
	suite.Require().NoError(err)
	suite.NotNil(suite.collection)
}

func (suite *indexTestSuite) TearDownTest() {
	_ = suite.collection.Drop(context.TODO())
}

func (suite *indexTestSuite) TestIndexNone() {
	NewIndexTester().TestIndexes(suite.T(), suite.collection)
}

func (suite *indexTestSuite) TestIndexOne() {
	index1 := NewIndexDescription(true, "alpha")
	suite.Require().NoError(suite.access.Index(suite.collection, index1))
	NewIndexTester().TestIndexes(suite.T(), suite.collection, index1)
}

func (suite *indexTestSuite) TestIndexTwo() {
	index1 := NewIndexDescription(true, "alpha")
	index2 := NewIndexDescription(true, "bravo")
	suite.Require().NoError(suite.access.Index(suite.collection, index1))
	suite.Require().NoError(suite.access.Index(suite.collection, index2))
	NewIndexTester().TestIndexes(suite.T(), suite.collection, index1, index2)
}

func (suite *indexTestSuite) TestIndexThree() {
	index1 := NewIndexDescription(true, "alpha")
	index2 := NewIndexDescription(true, "bravo")
	index3 := NewIndexDescription(true, "alpha", "bravo")
	suite.Require().NoError(suite.access.Index(suite.collection, index1))
	suite.Require().NoError(suite.access.Index(suite.collection, index2))
	suite.Require().NoError(suite.access.Index(suite.collection, index3))
	NewIndexTester().TestIndexes(suite.T(), suite.collection, index1, index2, index3)
}

func (suite *indexTestSuite) TestIndexFinisher() {
	index := NewIndexDescription(true, "alpha", "bravo")
	collection, err := suite.access.Collection(context.TODO(), "test-index-finisher-collection",
		testValidatorJSON, index.Finisher())
	suite.Require().NoError(err)
	suite.NotNil(collection)
	NewIndexTester().TestIndexes(suite.T(), collection, index)
}
