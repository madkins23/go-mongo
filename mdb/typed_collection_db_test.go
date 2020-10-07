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

func (suite *typedTestSuite) TestCreateDuplicate() {
	tk := &TestKey{
		Alpha: "two",
		Bravo: 2,
	}
	ti := &testItem{
		TestKey: *tk,
		Charlie: "Two is too much",
	}
	err := suite.typed.Create(ti)
	suite.Require().NoError(err)
	item, err := suite.typed.Find(tk.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	err = suite.typed.Create(ti)
	suite.Require().Error(err)
	suite.Require().True(suite.access.Duplicate(err))
}

func (suite *typedTestSuite) TestFindNone() {
	tk := &TestKey{
		Alpha: "beast",
		Bravo: 666,
	}
	item, err := suite.typed.Find(tk.Filter())
	suite.Require().Error(err)
	suite.True(suite.typed.NotFound(err))
	suite.Nil(item)
}

func (suite *typedTestSuite) TestCreateFindDelete() {
	tk := &TestKey{
		Alpha: "one",
		Bravo: 1,
	}
	ti := &testItem{
		TestKey: *tk,
		Charlie: "One is the loneliest number",
	}
	err := suite.typed.Create(ti)
	suite.Require().NoError(err)
	item, err := suite.typed.Find(tk.Filter())
	suite.Require().NoError(err)
	suite.NotNil(item)
	cacheKey := tk.CacheKey()
	suite.NotEmpty(cacheKey)
	err = suite.typed.Delete(tk.Filter(), false)
	suite.Require().NoError(err)
	noItem, err := suite.typed.Find(tk.Filter())
	suite.Require().Error(err)
	suite.True(suite.typed.NotFound(err))
	suite.Nil(noItem)
	err = suite.typed.Delete(tk.Filter(), false)
	suite.Require().Error(err)
	err = suite.typed.Delete(tk.Filter(), true)
	suite.Require().NoError(err)
}
