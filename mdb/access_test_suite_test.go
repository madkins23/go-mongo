package mdb

import (
	"github.com/stretchr/testify/suite"
)

const AccessTestDBname = "db-test"

type AccessTestSuite struct {
	suite.Suite
	access *Access
}

func (suite *AccessTestSuite) Access() *Access {
	return suite.access
}

func (suite *AccessTestSuite) SetupSuite() {
	suite.SetupSuiteConfig(nil)
}

func (suite *AccessTestSuite) SetupSuiteConfig(config *Config) {
	var err error
	suite.access, err = Connect(AccessTestDBname, config)
	suite.Require().NoError(err, "connect to mongo")
	suite.access.Info("Suite setup")
}

func (suite *AccessTestSuite) TearDownSuite() {
	suite.access.Info("Suite teardown")
	suite.NoError(suite.access.Database().Drop(suite.access.Context()), "drop test database")
	suite.NoError(suite.access.Disconnect(), "disconnect from mongo")
}

// ConnectCollection connects to the specified collection and adds any provided indexes
// as necessary in a SetupSuite() with test checks so that any errors blow up the test.
func (suite *AccessTestSuite) ConnectCollection(
	definition *CollectionDefinition, indexDescriptions ...*IndexDescription) *Collection {
	collection, err := ConnectCollection(suite.access, definition)
	suite.Require().NoError(err)
	suite.NotNil(collection)
	suite.Require().NoError(collection.DeleteAll())
	for _, indexDescription := range indexDescriptions {
		suite.Require().NoError(suite.access.Index(collection, indexDescription))
	}
	return collection
}

// ConnectTypedCollectionHelper is similar to AccessTestSuite.ConnectionCollection().
// Go doesn't support generic methods so this can't be a method on AccessTestSuite.
func ConnectTypedCollectionHelper[T any](
	suite *AccessTestSuite, definition *CollectionDefinition, indexDescriptions ...*IndexDescription) *TypedCollection[T] {
	collection, err := ConnectTypedCollection[T](suite.access, definition)
	suite.Require().NoError(err)
	suite.NotNil(collection)
	suite.Require().NoError(collection.DeleteAll())
	for _, indexDescription := range indexDescriptions {
		suite.Require().NoError(suite.access.Index(&collection.Collection, indexDescription))
	}
	return collection
}
