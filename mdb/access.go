package mdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Access encapsulates database connection.
type Access struct {
	client   *mongo.Client
	database *mongo.Database
	config   Config
}

var (
	// Default connection URL if not provided in Connect().
	DefaultURL = "mongodb://localhost:27017"

	// Default timeout for the initial connect.
	DefaultConnectTimeout = 10 * time.Second

	// Default timeout for the disconnect.
	DefaultDisconnectTimeout = 10 * time.Second

	// Default timeout for the ping to make sure the connection is up.
	DefaultPingTimeout = 2 * time.Second

	// Default timeout for collection access.
	DefaultCollectionTimeout = time.Second

	// Default timeout for index access.
	DefaultIndexTimeout = 5 * time.Second
)

// Config items for Mongo DB connection.
type Config struct {
	// Base context for use in calls to Mongo.
	Ctx context.Context

	// Mongo URL.
	URL string

	Timeout
}

// Timeout settings for Mongo DB access.
// TODO: should this be a map[string]time.Duration instead?
//  Use const to define the strings.
type Timeout struct {
	// Timeout for the initial connect.
	Connect time.Duration

	// Timeout for the disconnect.
	Disconnect time.Duration

	// Timeout for the ping to make sure the connection is up.
	Ping time.Duration

	// Timeout for collection access.
	Collection time.Duration

	// Timeout for indexes.
	Index time.Duration
}

// Connect to Mongo DB and return Access object.
// If the ctxt is nil it will be provided as context.Background().
// If the url is empty it will be set to mdb.DefaultURL.
// If the dbName is empty it will be set to mdb.DefaultDatabase.
func Connect(dbName string, config *Config) (*Access, error) {
	config = fixConfig(config)
	ctx, cancel := context.WithTimeout(config.Ctx, config.Timeout.Connect)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.URL))
	if err != nil {
		return nil, fmt.Errorf("unable to connect mongo server: %w", err)
	}

	access := &Access{
		client:   client,
		database: client.Database(dbName),
		config:   *config,
	}

	if err = access.Ping(); err != nil {
		return nil, err
	}

	access.Info("Connected to MongoDB database " + access.database.Name())

	return access, nil
}

// ConnectOrPanic connects to Mongo DB and returns Access object or panics on error.
func ConnectOrPanic(dbName string, config *Config) *Access {
	access, err := Connect(dbName, config)
	if err != nil {
		panic(err)
	}

	return access
}

// Disconnect Mongo DB client.
// Provided for use in defer statements.
func (a *Access) Disconnect() error {
	ctx, cancel := a.ContextWithTimeout(a.config.Timeout.Disconnect)
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

// Context returns the base context for the object.
func (a *Access) Context() context.Context {
	return a.config.Ctx
}

// Context returns the base context for the object with the specified timeout.
func (a *Access) ContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(a.config.Ctx, timeout)
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
func (a *Access) Ping() error {
	ctx, cancel := a.ContextWithTimeout(a.config.Timeout.Ping)
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

func fixConfig(config *Config) *Config {
	if config == nil {
		config = &Config{}
	}

	if config.Ctx == nil {
		config.Ctx = context.Background()
	}

	if config.URL == "" {
		config.URL = DefaultURL
	}

	if config.Timeout.Connect == 0 {
		config.Timeout.Connect = DefaultConnectTimeout
	}

	if config.Timeout.Disconnect == 0 {
		config.Timeout.Disconnect = DefaultDisconnectTimeout
	}

	if config.Timeout.Ping == 0 {
		config.Timeout.Ping = DefaultPingTimeout
	}

	if config.Timeout.Collection == 0 {
		config.Timeout.Collection = DefaultCollectionTimeout
	}

	if config.Timeout.Index == 0 {
		config.Timeout.Index = DefaultIndexTimeout
	}

	return config
}
