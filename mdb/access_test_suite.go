// +build database

package mdb

// Would prefer to name this file ending in _test.go
//  so that it won't be included in generated code
//  but then it can't be referenced from other packages
//  so it can't be used (as designed) in tests in other packages.

import (
	"context"

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
	// TODO: custom code to connect and drop database if it already exists

	var err error
	suite.access, err = Connect(nil, "", "db-test")
	suite.Require().NoError(err, "connect to mongo")
}

func (suite *AccessTestSuite) TearDownSuite() {
	suite.NoError(suite.access.database.Drop(context.TODO()), "drop test database")
	suite.NoError(suite.access.Disconnect(), "disconnect from mongo")
}
