// +build database

package mdb

import (
	"context"
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
	suite.NoError(suite.access.Ping(context.TODO()))
}
