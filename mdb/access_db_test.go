// +build database

package mdb

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type accessTestSuite struct {
	AccessTestSuite
}

func TestAccessSuite(t *testing.T) {
	suite.Run(t, new(accessTestSuite))
}

func (suite *accessTestSuite) TestPing() {
	suite.NoError(suite.access.Ping())
}

func (suite *accessTestSuite) TestContext() {
	suite.Require().NotNil(suite.access.Context())
}
