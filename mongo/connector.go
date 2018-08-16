package mongo

import "github.com/globalsign/mgo"

type Connector interface {
	GetCollection(name string) Collection
}

type connector struct {
	db *mgo.Database
}

const (
	collectionName =  "topics"
)

func New(connectionURL, databaseName string) (Connector, error) {
	session, err := mgo.Dial(connectionURL)
	if err != nil {
		return nil, err
	}
	return &connector{db: session.DB(databaseName)},nil
}

func (s *connector) GetCollection(name string) Collection {
	return &collection{
		collectionName: name,
		collection: s.db.C(name),
	}
}