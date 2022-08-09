package mdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Access encapsulates database connection.
type Access struct {
	client   *mongo.Client
	database *mongo.Database
	config   Config
}

var (
	// DefaultURI is the default connection URI if not provided in Config.Options.
	DefaultURI = "mongodb://localhost:27017"

	// DefaultLogInfoFn is the default info logging function.
	DefaultLogInfoFn = func(msg string) {
		fmt.Printf("MDB: %s\n", msg)
	}

	// DefaultConnectTimeout is the default timeout for the initial connect.
	DefaultConnectTimeout = 10 * time.Second

	// DefaultDisconnectTimeout is the default timeout for the disconnect.
	DefaultDisconnectTimeout = 10 * time.Second

	// DefaultPingTimeout is the default timeout for the ping to make sure the connection is up.
	DefaultPingTimeout = 2 * time.Second

	// DefaultCollectionTimeout is the default timeout for collection access.
	DefaultCollectionTimeout = time.Second

	// DefaultIndexTimeout is the default timeout for index access.
	DefaultIndexTimeout = 5 * time.Second
)

// Config items for Mongo DB connection.
type Config struct {
	// Base context for use in calls to Mongo.
	Ctx context.Context

	// Mongo options.
	Options *options.ClientOptions

	// Optional BSON codec registry for handling special types.
	Registry *bsoncodec.Registry

	// Logging function for information messages may be overridden.
	LogInfoFn func(msg string)
	// Errors should bubble up and be handled by client code.

	Timeout
}

// Timeout settings for Mongo DB access.
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

var ErrNoDbName = errors.New("no database name")

// Connect to Mongo DB and return Access object.
// If the ctxt is nil it will be provided as context.Background().
// If the url is empty it will be set to mdb.DefaultURI.
func Connect(dbName string, config *Config) (*Access, error) {
	if dbName == "" {
		return nil, ErrNoDbName
	}

	config = fixConfig(config)
	ctx, cancel := context.WithTimeout(config.Ctx, config.Timeout.Connect)
	defer cancel()

	client, err := mongo.Connect(ctx, config.Options)
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

// ContextWithTimeout returns the base context for the object with the specified timeout.
func (a *Access) ContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(a.config.Ctx, timeout)
}

// Database returns the Mongo database object.
func (a *Access) Database() *mongo.Database {
	return a.database
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
	a.config.LogInfoFn(msg)
}

func fixConfig(config *Config) *Config {
	if config == nil {
		config = &Config{}
	}

	if config.Ctx == nil {
		config.Ctx = context.Background()
	}

	if config.Options == nil {
		config.Options = options.Client()
		if config.Options.GetURI() == "" {
			config.Options.ApplyURI(DefaultURI)
		}
	}

	if config.LogInfoFn == nil {
		config.LogInfoFn = DefaultLogInfoFn
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

////////////////////////////////////////////////////////////////////////////////

var errMissingCollectionName = errors.New("no collection name argument")

// CollectionExists checks to see if a specific collection already exists.
func (a *Access) CollectionExists(name string) (bool, error) {
	if name == "" {
		return false, errMissingCollectionName
	}

	ctx, cancel := a.ContextWithTimeout(a.config.Timeout.Collection)
	defer cancel()
	names, err := a.database.ListCollectionNames(ctx, bson.M{"name": name})
	if err != nil {
		return false, fmt.Errorf("getting collection names: %w", err)
	}

	exists := false
	for _, collName := range names {
		if collName == name {
			exists = true
			break
		}
	}

	return exists, nil
}

// CollectionFinisher provides a way to add special processing when creating a collection.
type CollectionFinisher func(access *Access, collection *Collection) error

// Collection acquires the named collection, creating it if necessary.
func (a *Access) Collection(
	ctx context.Context, collectionName string, validatorJSON string, finishers ...CollectionFinisher) (*Collection, error) {
	if collectionName == "" {
		return nil, errMissingCollectionName
	}

	if exists, err := a.CollectionExists(collectionName); err != nil {
		return nil, fmt.Errorf("does collection '%s' exist: %w", collectionName, err)
	} else if exists {
		// Collection already exists, just return it.
		return &Collection{Access: a, Collection: a.database.Collection(collectionName)}, nil
	}

	// Add option for validator JSON if it is provided.
	opts := make([]*options.CreateCollectionOptions, 0)
	if validatorJSON != "" {
		var validator interface{}
		if err := bson.UnmarshalExtJSON([]byte(validatorJSON), false, &validator); err != nil {
			return nil, fmt.Errorf("unmarshal validator for collection: %w", err)
		}
		opts = append(opts, &options.CreateCollectionOptions{Validator: validator})
	}

	// Create collection.

	createCtx, cancel := a.ContextWithTimeout(a.config.Timeout.Collection)
	defer cancel()
	err := a.database.CreateCollection(createCtx, collectionName, opts...)
	if err != nil {
		if cmdErr, ok := err.(mongo.CommandError); !ok || !cmdErr.HasErrorLabel("NamespaceExists") {
			return nil, fmt.Errorf("create collection: %w", err)
		}
	}
	if ctx == nil {
		ctx = a.Context()
	}
	collection := &Collection{
		Access:     a,
		Collection: a.database.Collection(collectionName),
		ctx:        ctx,
	}
	a.Info("Created collection " + collection.Name())

	// Run finishers on the collection.
	for i, finisher := range finishers {
		if err = finisher(a, collection); err != nil {
			return nil, fmt.Errorf("collection finisher #%d: %w", i, err)
		}
	}

	return collection, nil
}

////////////////////////////////////////////////////////////////////////////////
// Functions to check for specific, known errors.

// IsDuplicate checks to see if the specified error is for attempting to create a duplicate document.
func IsDuplicate(err error) bool {
	if err == nil {
		return false
	}

	var e mongo.WriteException
	if errors.As(err, &e) {
		for _, we := range e.WriteErrors {
			if we.Code == 11000 {
				return true
			}
		}
	}

	return false
}

// IsNotFound checks an error condition to see if it matches the underlying database "not found" error.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, mongo.ErrNoDocuments)
}

// IsValidationFailure checks to see if the specified error is for a validation failure.
func IsValidationFailure(err error) bool {
	if err == nil {
		return false
	}

	var e mongo.WriteException
	if errors.As(err, &e) {
		for _, we := range e.WriteErrors {
			if we.Code == 121 {
				return true
			}
		}
	}

	return false
}
