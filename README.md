# mdb
--
    import "."

Package mdb provides infrastructure for using Mongo from Go.

This README file was generated using github.com/robertkrimen/godocdown

Use the Connect() function to connect to the DB and return an Access object. The
Access object provides access to the Mongo DB and some common functionality. The
dbName is required for Connect(), an optional pointer to an Config struct can be
used to provide additional parameters for connecting to the DB. If these are not
provided they are filled in from various default global variables which are
visible and may be changed.

The Access object provides a Disconnect() method suitable for use with defer.

In addition, the Access object can be used to construct collections. The
Collection() call takes a collection name, an optional validation JSON string,
and optional list of "finisher" functions intended to create indices or
otherwise configure the collection after it is created. The Index() call is used
to add an index to a collection.

The CachedCollection object provides a caching layer for mostly static tables.
Implementing the various provided interfaces in table record objects allows them
to be created, deleted, and found by the CachedCollection object.

The AccessTestSuite struct is provided to wrap database connect/disconnect for
use in tests that actually hit the database. The use of '+build database'
separates these so that they are only run when using 'go test -tags database',
without this tag only unit tests are run.

The IndexTester supports verifying that a specific index has been added to a
collection.

## Usage

```go
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
```

#### type Access

```go
type Access struct {
}
```

Access encapsulates database connection.

#### func  Connect

```go
func Connect(dbName string, config *Config) (*Access, error)
```
Connect to Mongo DB and return Access object. If the ctxt is nil it will be
provided as context.Background(). If the url is empty it will be set to
mdb.DefaultURL. If the dbName is empty it will be set to mdb.DefaultDatabase.

#### func  ConnectOrPanic

```go
func ConnectOrPanic(dbName string, config *Config) *Access
```
ConnectOrPanic connects to Mongo DB and returns Access object or panics on
error.

#### func (*Access) Client

```go
func (a *Access) Client() *mongo.Client
```
Client returns the Mongo client object.

#### func (*Access) Collection

```go
func (a *Access) Collection(collectionName string, validatorJSON string, finishers ...CollectionFinisher) (*mongo.Collection, error)
```
Collection acquires the named collection, creating it if necessary.

#### func (*Access) CollectionExists

```go
func (a *Access) CollectionExists(name string) (bool, error)
```
CollectionExists checks to see if a specific collection already exists.

#### func (*Access) Context

```go
func (a *Access) Context() context.Context
```
Context returns the base context for the object.

#### func (*Access) ContextWithTimeout

```go
func (a *Access) ContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc)
```
Context returns the base context for the object with the specified timeout.

#### func (*Access) Database

```go
func (a *Access) Database() *mongo.Database
```
Client returns the Mongo database object.

#### func (*Access) Disconnect

```go
func (a *Access) Disconnect() error
```
Disconnect Mongo DB client. Provided for use in defer statements.

#### func (*Access) DisconnectOrPanic

```go
func (a *Access) DisconnectOrPanic()
```
DisconnectOrPanic disconnects the Mongo DB client or panics on error. Provided
for use in defer statements.

#### func (*Access) Duplicate

```go
func (a *Access) Duplicate(err error) bool
```

#### func (*Access) Index

```go
func (a *Access) Index(collection *mongo.Collection, description *IndexDescription) error
```

#### func (*Access) Info

```go
func (a *Access) Info(msg string)
```
Info prints a simple message in the format MDB: <msg>. This is used for a few
calls within the Access code. It may be overridden to use another logger or to
block these messages.

#### func (*Access) NotFound

```go
func (a *Access) NotFound(err error) bool
```
NotFound checks an error condition to see if it matches the underlying database
"not found" error.

#### func (*Access) Ping

```go
func (a *Access) Ping() error
```
Ping executes a ping against the Mongo server. This is separated from Connect()
so that it can be overridden if necessary.

#### type AccessTestSuite

```go
type AccessTestSuite struct {
	suite.Suite
}
```


#### func (*AccessTestSuite) Access

```go
func (suite *AccessTestSuite) Access() *Access
```

#### func (*AccessTestSuite) SetupSuite

```go
func (suite *AccessTestSuite) SetupSuite()
```

#### func (*AccessTestSuite) TearDownSuite

```go
func (suite *AccessTestSuite) TearDownSuite()
```

#### type Cacheable

```go
type Cacheable interface {
	Searchable
	ExpireAfter(time.Duration)
	Expired() bool
	InitFrom(stub bson.M) error
}
```

Cacheable must be searchable and able to be recreated and expired.

#### type CachedCollection

```go
type CachedCollection struct {
	*Access
	*mongo.Collection
}
```

CachedCollection Mongo-stored objects so that the same object is always
returned. This is most useful for objects that change rarely.

#### func  NewCache

```go
func NewCache(
	access *Access, collection *mongo.Collection,
	ctx context.Context, example interface{}, expireAfter time.Duration) *CachedCollection
```

#### func (*CachedCollection) Create

```go
func (c *CachedCollection) Create(item Storable) error
```
Create object in DB but not cache.

#### func (*CachedCollection) Delete

```go
func (c *CachedCollection) Delete(item Searchable, idempotent bool) error
```
Delete object in cache and DB.

#### func (*CachedCollection) Find

```go
func (c *CachedCollection) Find(searchFor Searchable) (Cacheable, error)
```
Find a cacheable object in either cache or database.

#### type CollectionFinisher

```go
type CollectionFinisher func(access *Access, collection *mongo.Collection) error
```

CollectionFinisher provides a way to add special processing when creating a
collection.

#### type Config

```go
type Config struct {
	// Base context for use in calls to Mongo.
	Ctx context.Context

	// Mongo URL.
	URL string

	Timeout
}
```

Config items for Mongo DB connection.

#### type IndexDescription

```go
type IndexDescription struct {
}
```


#### func  NewIndexDescription

```go
func NewIndexDescription(unique bool, keys ...string) *IndexDescription
```
Create new index description.

#### func (*IndexDescription) AsBSON

```go
func (id *IndexDescription) AsBSON() bson.D
```

#### type IndexTester

```go
type IndexTester []indexDatum
```

IndexTester provides a utility for verifying index creation.

#### func  NewIndexTester

```go
func NewIndexTester() IndexTester
```

#### func (IndexTester) TestIndexes

```go
func (it IndexTester) TestIndexes(t *testing.T, collection *mongo.Collection, descriptions ...*IndexDescription)
```

#### type Searchable

```go
type Searchable interface {
	CacheKey() (string, error)
	Filter() bson.D
}
```

Searchable may be used just for searching for a cached item. This supports keys
that are not complete items.

#### type Storable

```go
type Storable interface {
	Document() bson.M
}
```

Storable must be able to generate a Mongo document. This supports "stub" objects
used just to add items.

#### type Timeout

```go
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
```

Timeout settings for Mongo DB access.
