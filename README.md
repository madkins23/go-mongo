# go-mongo

Provides some potentially useful functionality wrapped around the
[MongoDB `go` driver](https://github.com/mongodb/mongo-go-driver).

![GitHub](https://img.shields.io/github/license/madkins23/go-mongo)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/madkins23/go-mongo)

### Caveats

You are more than welcome to use this software as is but these are
utility packages constructed by the author for use in personal projects.
The author makes occasional changes and attempts to follow proper versioning and release protocols,
however this code should not be considered production quality or maintained.

*Consider copying the code into your own project and modifying to fit your need.*

See the [source](https://github.com/madkins23/go-mongo)
or [godoc](https://godoc.org/github.com/madkins23/go-mongo) for documentation.

## Package `mdb`

Provides infrastructure for connecting to a Mongo
database and collections.
Collections can be untyped (i.e. `interface{}`) or typed using generics.
Support is provided for defining indexes at the time of collection creation.
Collections support a simplified set of functionality and the basic
`mongo.collection` functionality is always accessible.

## Package `mdbson`

Supports marshaling and unmarshaling structs with fields that are interfaces.[^1]

[^1]: This is implemented the same way as [`madkins23/go-serial`](https://github.com/madkins23/go-serial)
