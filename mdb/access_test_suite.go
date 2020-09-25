// +build database

package mdb

// Would prefer to name this file ending in _test.go
//  so that it won't be included in generated code
//  but then it can't be referenced from other packages
//  so it can't be used (as designed) in tests in other packages.

import (
	"github.com/stretchr/testify/suite"
)

type AccessTestSuite struct {
	suite.Suite
	access *Access
}

func (suite *AccessTestSuite) Access() *Access {
	return suite.access
}

func (suite *AccessTestSuite) SetupSuite() {
	var err error
	suite.access, err = Connect("db-test", nil)
	suite.Require().NoError(err, "connect to mongo")
	suite.access.Info("Suite setup")
}

func (suite *AccessTestSuite) TearDownSuite() {
	suite.access.Info("Suite teardown")
	suite.NoError(suite.access.database.Drop(suite.access.Context()), "drop test database")
	suite.NoError(suite.access.Disconnect(), "disconnect from mongo")
}
