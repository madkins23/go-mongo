package mdb

import "github.com/madkins23/go-mongo/test"

var (
	testCollection = &CollectionDefinition{
		name: "test-collection",
	}
	testCollectionValidation = &CollectionDefinition{
		name:           "test-collection-validation",
		validationJSON: test.SimpleValidatorJSON,
	}
	testCollectionStringValues = &CollectionDefinition{
		name: "test-collection-string-values",
	}
	testCollectionWrapped = &CollectionDefinition{
		name: "test-collection-wrapped",
	}
	testCollectionIndexFinisher = &CollectionDefinition{
		name:           "test-collection-index-finisher",
		validationJSON: test.SimpleValidatorJSON,
	}
)
