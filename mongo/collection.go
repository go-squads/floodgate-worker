package mongo

import "github.com/globalsign/mgo"

type Collection interface {
	Insert(data interface{}) error
}

type collection struct {
	collectionName string
	collection     *mgo.Collection
}

func (s *collection) Insert(data interface{}) error {
	err := s.collection.Insert(data)
	return err
}
