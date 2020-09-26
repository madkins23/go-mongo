// Package mdb provides infrastructure for using Mongo from Go.
// This package uses the zerolog logging package.
//
// Use the Connect() function to connect to the DB and return an Access object.
// The Access object provides access to the Mongo DB and some common functionality.
// The dbName is required for Connect(), an optional pointer to an Config struct
// can be used to provide additional parameters for connecting to the DB.
// If these are not provided they are filled in from various default global variables
// which are visible and may be changed.
//
// The Access object provides a Disconnect() method suitable for use with defer.
//
// In addition, the Access object can be used to construct collections.
// The Collection() call takes a collection name, an optional validation JSON string,
// and optional list of "finisher" functions intended to create indices
// or otherwise configure the collection after it is created.
//
// The AccessTestSuite struct is provided to wrap database connect/disconnect
// for use in tests that actually hit the database.
// The use of '+build database' separates these so that they are only run
// when using 'go test -tags database', without this tag only unit tests are run.
package mdb

import "errors"

var (
	NotYetImplemented = errors.New("Not yet implemented")
)
