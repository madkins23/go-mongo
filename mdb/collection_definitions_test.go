package mdb

var (
	testCollection = &CollectionDefinition{
		Name: "test-collection",
	}
	testCollectionValidation = &CollectionDefinition{
		Name:           "test-collection-validation",
		ValidationJSON: SimpleValidatorJSON,
	}
	testCollectionStringValues = &CollectionDefinition{
		Name: "test-collection-string-values",
	}
	testCollectionWrapped = &CollectionDefinition{
		Name: "test-collection-wrapped",
	}
)
