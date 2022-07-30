package mdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestConnectOrPanic(t *testing.T) {
	// Cause a failure by using a bad URI.
	opts := options.Client()
	opts.ApplyURI("bad URI")
	assert.Panics(t, func() {
		ConnectOrPanic("noSuchDB", &Config{Options: opts})
	}, "TestConnectOrPanic did not panic")
}

// TODO(mAdkins): is there a way to force a disconnect failure?
//  Forcing the disconnect timeout to zero doesn't work.
//func TestDisconnectOrPanic(t *testing.T) {
//	oldTimeout := DefaultDisconnectTimeout
//	DefaultDisconnectTimeout = 0
//	defer func() { DefaultDisconnectTimeout = oldTimeout }()
//	access, err := Connect(AccessTestDBname, nil)
//	require.NoError(t, err)
//	require.NotNil(t, access)
//	assert.Panics(t, func() {
//		access.DisconnectOrPanic()
//	}, "TestConnectOrPanic did not panic")
//}
