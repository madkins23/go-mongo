//go:build database
// +build database

package mdb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
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
	suite.Require().True(suite.access.IsDuplicate(err))
}

func (suite *typedTestSuite) TestFindNone() {
	item, err := suite.typed.Find(testKeyOfTheBeast.Filter())
	suite.Require().Error(err)
	suite.True(suite.typed.IsNotFound(err))
	suite.Nil(item)
}

func (suite *typedTestSuite) TestCreateFindDelete() {
	err := suite.typed.Create(testItem2)
	suite.Require().NoError(err)
	item, err := suite.typed.Find(testItem2.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	ti, ok := item.(*testItem)
	suite.Require().True(ok)
	suite.True(ti.Realized)
	cacheKey := testItem2.CacheKey()
	suite.NotEmpty(cacheKey)
	err = suite.typed.Delete(testItem2.Filter(), false)
	suite.Require().NoError(err)
	noItem, err := suite.typed.Find(testItem2.Filter())
	suite.Require().Error(err)
	suite.True(suite.typed.IsNotFound(err))
	suite.Nil(noItem)
	err = suite.typed.Delete(testItem2.Filter(), false)
	suite.Require().Error(err)
	err = suite.typed.Delete(testItem2.Filter(), true)
	suite.Require().NoError(err)
}

func (suite *typedTestSuite) TestIterate() {
	suite.Require().NoError(suite.typed.Create(testItem1))
	suite.Require().NoError(suite.typed.Create(testItem2))
	suite.Require().NoError(suite.typed.Create(testItem3))
	count := 0
	alpha := []string{}
	suite.NoError(suite.typed.Iterate(bson.D{}, func(item interface{}) error {
		ti, ok := item.(*testItem)
		suite.Require().True(ok)
		suite.True(ti.Realized)
		alpha = append(alpha, ti.Alpha)
		count++
		return nil
	}))
	suite.Equal(3, count)
	suite.Equal([]string{"one", "two", "three"}, alpha)
}

func (suite *typedTestSuite) TestIterateFiltered() {
	suite.Require().NoError(suite.typed.Create(testItem1))
	suite.Require().NoError(suite.typed.Create(testItem2))
	suite.Require().NoError(suite.typed.Create(testItem3))
	count := 0
	alpha := []string{}
	suite.NoError(suite.typed.Iterate(bson.D{bson.E{Key: "bravo", Value: 2}}, func(item interface{}) error {
		ti, ok := item.(*testItem)
		suite.Require().True(ok)
		suite.True(ti.Realized)
		alpha = append(alpha, ti.Alpha)
		count++
		return nil
	}))
	suite.Equal(1, count)
	suite.Equal([]string{"two"}, alpha)
}
