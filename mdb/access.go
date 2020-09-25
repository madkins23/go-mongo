package mdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	// Default database name if not provided in Connect().
	DefaultDatabase = "test"

	// Default connection URL if not provided in Connect().
	DefaultURL = "mongodb://localhost:27017"
)

var (
	// Timeout for the initial connect.
	ConnectTimeout = 10 * time.Second

	// Timeout for the disconnect.
	DisconnectTimeout = 10 * time.Second

	// Timeout for the ping to make sure the connection is up.
	PingTimeout = 2 * time.Second
)

// Access encapsulates database connection.
type Access struct {
	client   *mongo.Client
	database *mongo.Database
}

// Connect to Mongo DB and return Access object.
// If the ctxt is nil it will be provided as context.Background().
// If the url is empty it will be set to mdb.DefaultURL.
// If the dbName is empty it will be set to mdb.DefaultDatabase.
func Connect(ctxt context.Context, url string, dbName string) (*Access, error) {
	if ctxt == nil {
		ctxt = context.Background()
	}

	if url == "" {
		url = DefaultURL
	}

	if dbName == "" {
		dbName = DefaultDatabase
	}

	ctx, cancel := context.WithTimeout(ctxt, ConnectTimeout)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		return nil, fmt.Errorf("unable to connect mongo server: %w", err)
	}

	access := &Access{
		client:   client,
		database: client.Database(dbName),
	}

	if err = access.Ping(ctxt); err != nil {
		return nil, err
	}

	access.Info("Connected to MongoDB database " + access.database.Name())

	return access, nil
}

// ConnectOrPanic connects to Mongo DB and returns Access object or panics on error.
func ConnectOrPanic(ctxt context.Context, url string, dbName string) *Access {
	access, err := Connect(ctxt, url, dbName)
	if err != nil {
		panic(err)
	}

	return access
}

// Disconnect Mongo DB client.
// Provided for use in defer statements.
func (a *Access) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), DisconnectTimeout)
	defer cancel()
	if err := a.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("unable to disconnect mongo server: %w", err)
	}

	return nil
}

// DisconnectOrPanic disconnects the Mongo DB client or panics on error.
// Provided for use in defer statements.
func (a *Access) DisconnectOrPanic() {
	if err := a.Disconnect(); err != nil {
		panic(err)
	}
}

// Client returns the Mongo client object.
func (a *Access) Client() *mongo.Client {
	return a.client
}

// Client returns the Mongo database object.
func (a *Access) Database() *mongo.Database {
	return a.database
}

// NotFound checks an error condition to see if it matches the underlying database "not found" error.
func (a *Access) NotFound(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, mongo.ErrNoDocuments)
}

// Ping executes a ping against the Mongo server.
// This is separated from Connect() so that it can be overridden if necessary.
func (a *Access) Ping(ctxt context.Context) error {
	ctx, cancel := context.WithTimeout(ctxt, PingTimeout)
	defer cancel()
	err := a.client.Ping(ctx, readpref.Primary())
	if err != nil {
		return fmt.Errorf("unable to ping mongo server: %w", err)
	}

	return nil
}

// Info prints a simple message in the format MDB: <msg>.
// This is used for a few calls within the Access code.
// It may be overridden to use another logger or to block these messages.
func (a *Access) Info(msg string) {
	fmt.Printf("MDB: %s\n", msg)
}
