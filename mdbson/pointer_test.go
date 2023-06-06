package mdbson

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/madkins23/go-serial/pointer"
	"github.com/madkins23/go-serial/test"
)

type BsonPointerTestSuite struct {
	suite.Suite
	showSerialized bool
}

func (suite *BsonPointerTestSuite) SetupSuite() {
	if showSerialized, found := os.LookupEnv("GO-TYPE-SHOW-SERIALIZED"); found {
		var err error
		suite.showSerialized, err = strconv.ParseBool(showSerialized)
		suite.Require().NoError(err)
	}
	pointer.ClearTargetCache()
	pointer.ClearFinderCache()
	suite.Require().NoError(test.CachePets())
}

func TestBsonPointerSuite(t *testing.T) {
	suite.Run(t, new(BsonPointerTestSuite))
}

//////////////////////////////////////////////////////////////////////////

func (suite *BsonPointerTestSuite) TestPointer() {
	ptr := Point[*test.Pet](test.Lacey)
	suite.Assert().Equal(test.Lacey, ptr.Get())
	ptr.Set(test.Noah)
	suite.Assert().Equal(test.Noah, ptr.Get())
}

type animals struct {
	Cats []*Pointer[*test.Pet]
	Dog  *Pointer[*test.Pet]
}

func makeAnimals() *animals {
	return &animals{
		Cats: []*Pointer[*test.Pet]{
			Point[*test.Pet](test.Noah),
			Point[*test.Pet](test.Lacey),
			Point[*test.Pet](test.Orca),
		},
		Dog: Point[*test.Pet](test.Knight),
	}
}

func (suite *BsonPointerTestSuite) TestMarshalCycle() {
	start := makeAnimals()
	marshaled, err := bson.Marshal(start)
	suite.Require().NoError(err)
	suite.Require().NotNil(marshaled)

	finish := new(animals)
	suite.Require().NotNil(finish)
	suite.Require().NoError(bson.Unmarshal(marshaled, finish))
	if suite.showSerialized {
		fmt.Println("---------------------------")
		spew.Dump(finish)
	}

	suite.Require().Equal(start, finish)
}
