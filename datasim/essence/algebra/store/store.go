package store

import (
	"math/rand"
	"time"
)

func NewStore() *Store {
	return &Store{}
}

type Storer interface {
	Set(id EntityID, name EntityName, key AttributeName, value AttributeValue) error
	Get(id EntityID, name EntityName, strings AttrList) (interface{}, error)
	GetAttributes(name EntityName) AttrList
}


type Store struct{
}

type (
	EntityID       = string
	EntityName     = string
	AttributeName  = string
	AttributeValue = interface{}
	AttrList       []AttributeName
)

func (s *Store) Set(id EntityID, name EntityName, key AttributeName, value AttributeValue) error {
	//time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	time.Sleep(time.Duration(rand.NormFloat64() * 100) * time.Millisecond)
	return nil
}

//func (s *Store) Update(id EntityID, name EntityName, key AttributeName, value AttributeValue) {
//
//}

//func (s *Store) Delete(id EntityID, name EntityName, key AttributeName) {
//
//}

func (s *Store) Get(id EntityID, name EntityName, strings AttrList) (interface{}, error){
	return nil, nil
}

func (s *Store) GetAttributes(name EntityName) AttrList {
	return nil
}
