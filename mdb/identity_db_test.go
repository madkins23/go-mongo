package mdb

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
)

type identityTestSuite struct {
	AccessTestSuite
	collection *Collection
}

func TestIdentitySuite(t *testing.T) {
	suite.Run(t, new(identityTestSuite))
}

func (suite *identityTestSuite) SetupSuite() {
	suite.AccessTestSuite.SetupSuite()
	suite.collection = suite.ConnectCollection(testCollection)
}

func (suite *identityTestSuite) TearDownTest() {
	suite.NoError(suite.collection.DeleteAll())
}

const findable = "findable"

type identified struct {
	Identity `bson:"inline"`
	Text     string
}

func (suite *identityTestSuite) TestIndex() {
	ind := new(identified)
	suite.Require().NoError(suite.collection.Create(ind))
	ind.Text = findable
	suite.Require().NoError(suite.collection.Create(ind))
	found := suite.findStruct(bson.D{{"text", findable}})
	// Do we have an email?
	suite.Require().NotNil(found.ObjectID)
	suite.Require().NotNil(found.ID())
	suite.Equal(found.ObjectID, found.ID())
	suite.NotEmpty(found.ID())
	foundID := found.ID()
	// Use IDfilter() to find the item.
	found = suite.findStruct(found.IDfilter())
	suite.Require().NotNil(found.ID())
	suite.NotEmpty(found.ID())
	suite.Equal(foundID, found.ID())
}

func (suite *identityTestSuite) findStruct(filter bson.D) *identified {
	found, err := suite.collection.Find(filter)
	suite.Require().NoError(err)
	suite.Require().NotNil(found)
	// Convert bson.D nested structure to identified struct.
	var bytes []byte
	bytes, err = bson.Marshal(found)
	suite.Require().NoError(err)
	newInd := new(identified)
	suite.Require().NoError(bson.Unmarshal(bytes, newInd))
	return newInd
}
