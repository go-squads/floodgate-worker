package mongo

import "github.com/globalsign/mgo"

type Connector interface {
	GetCollection(name string) Collection
}

type connector struct {
	session *mgo.Session
	db *mgo.Database
}

func New(connectionURL, databaseName string) (Connector, error) {
	session, err := mgo.Dial(connectionURL)
	if err != nil {
		return nil, err
	}
	db := session.DB(databaseName)
	return &connector{session: session, db: db}, nil
}

func (s *connector) GetCollection(name string) Collection {
	return &collection{
		collectionName: name,
		collection:     s.db.C(name),
	}
}

func (s *connector)Close() {
	s.session.Close()
}