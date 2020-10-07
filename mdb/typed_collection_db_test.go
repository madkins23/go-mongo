// +build database

package mdb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type typedTestSuite struct {
	AccessTestSuite
	typed *TypedCollection
}

func TestTypedSuite(t *testing.T) {
	suite.Run(t, new(typedTestSuite))
}

func (suite *typedTestSuite) SetupSuite() {
	suite.AccessTestSuite.SetupSuite()
	collection, err := suite.access.Collection(context.TODO(), "test-collection", testValidatorJSON)
	suite.Require().NoError(err)
	suite.NotNil(collection)
	suite.Require().NoError(suite.access.Index(collection, NewIndexDescription(true, "alpha")))
	suite.typed = NewTypedCollection(collection, &testItem{})
}

func (suite *typedTestSuite) TearDownTest() {
	suite.NoError(suite.typed.DeleteAll())
}

func (suite *typedTestSuite) TestCreateDuplicate() {
	err := suite.typed.Create(testItem1)
	suite.Require().NoError(err)
	item, err := suite.typed.Find(testItem1.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	err = suite.typed.Create(testItem1)
	suite.Require().Error(err)
	suite.Require().True(suite.access.Duplicate(err))
}

func (suite *typedTestSuite) TestFindNone() {
	item, err := suite.typed.Find(testKeyOfTheBeast.Filter())
	suite.Require().Error(err)
	suite.True(suite.typed.NotFound(err))
	suite.Nil(item)
}

func (suite *typedTestSuite) TestCreateFindDelete() {
	err := suite.typed.Create(testItem2)
	suite.Require().NoError(err)
	item, err := suite.typed.Find(testItem2.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	cacheKey := testItem2.CacheKey()
	suite.NotEmpty(cacheKey)
	err = suite.typed.Delete(testItem2.Filter(), false)
	suite.Require().NoError(err)
	noItem, err := suite.typed.Find(testItem2.Filter())
	suite.Require().Error(err)
	suite.True(suite.typed.NotFound(err))
	suite.Nil(noItem)
	err = suite.typed.Delete(testItem2.Filter(), false)
	suite.Require().Error(err)
	err = suite.typed.Delete(testItem2.Filter(), true)
	suite.Require().NoError(err)
}
