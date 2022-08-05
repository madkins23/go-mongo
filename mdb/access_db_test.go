//go:build database

package mdb

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type accessTestSuite struct {
	AccessTestSuite
}

func (suite *accessTestSuite) TestConnectOrPanic() {
	// Cause a failure by using a bad URI.
	opts := options.Client()
	opts.ApplyURI("bad URI")
	suite.Panics(func() {
		ConnectOrPanic("noSuchDB", &Config{Options: opts})
	}, "TestConnectOrPanic did not panic")
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

func (suite *accessTestSuite) DontTestDisconnectOrPanic() {
	// TODO(mAdkins): is there a way to force a disconnect failure?
	//  Forcing the disconnect timeout to zero doesn't work.
	oldTimeout := DefaultDisconnectTimeout
	DefaultDisconnectTimeout = 0
	defer func() { DefaultDisconnectTimeout = oldTimeout }()
	access, err := Connect(AccessTestDBname, nil)
	suite.Require().NoError(err)
	suite.Require().NotNil(access)
	suite.Panics(func() {
		access.DisconnectOrPanic()
	}, "TestConnectOrPanic did not panic")
}
