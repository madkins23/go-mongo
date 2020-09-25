// Package mdb provides infrastructure for using Mongo from Go.
// This package uses the zerolog logging package.
//
// The Access struct contains the current Mongo client and database objects.
// It is returned from the Connect() function which also pings the database.
// Visible variables can be used to change default configuration and timeouts.
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
