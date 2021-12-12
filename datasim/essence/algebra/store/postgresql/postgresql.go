package postgresql

import (
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/widmogrod/software-architecture-playground/datasim/essence/algebra/store"
	"reflect"
	"strings"
)

func NewStore(db *sql.DB, relation store.PrimaryWithMany) *Store {
	return &Store{
		db:    db,
		shape: relation,
	}
}

var (
	_ store.Storer      = &Store{}
	_ store.TypeVisitor = &Store{}
)

type Store struct {
	db    *sql.DB
	shape store.PrimaryWithMany
}

func (s *Store) VisitInt(x store.Int) interface{} {
	return "int"
}

func (s *Store) VisitString(x store.String) interface{} {
	return "text"
}

func (s *Store) VisitDateTime(x store.DateTime) interface{} {
	return "timestamp with time zone"
}

func (s *Store) Set(id store.EntityID, name store.EntityName, key store.AttributeName, value store.AttributeValue) error {
	entity, found := FindEntity(s.shape, name)
	if !found {
		return fmt.Errorf("entity with name %s does not exits in shape", name)
	}

	r := s.getColumnName(entity, s.shape.Primary, key)
	pk := s.getPrimaryKey(entity, s.shape.Primary)

	b := sq.
		Insert(r).
		Columns(
			pk,
			"key",
			"value",
		).
		Values(id, key, value).
		PlaceholderFormat(sq.Dollar).
		Suffix(fmt.Sprintf("on conflict(%s, key) DO UPDATE SET value = excluded.value, version = %s.version + 1", pk, r))

	_, err := b.RunWith(s.db).Exec()
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Get(id store.EntityID, name store.EntityName, attrList store.AttrList) (interface{}, error) {
	entity, found := FindEntity(s.shape, name)
	if !found {
		return nil, fmt.Errorf("entity with name %s does not exits in shape", name)
	}

	pk := s.getPrimaryKey(entity, s.shape.Primary)

	values := map[string]interface{}{}
	typesScan := map[string]bool{}

	for _, attr := range entity.Attributes {
		switch a := attr.(type) {
		case store.Attribute:
			if !HasAttribute(attrList, a.Name) {
				continue
			}

			typ := store.MapType(a.Type, s).(string)
			if _, ok := typesScan[typ]; ok {
				continue
			}
			typesScan[typ] = true

			tableName := s.getColumnName(entity, s.shape.Primary, a.Name)
			b := sq.
				Select("key", "value").
				From(tableName).
				Where(sq.Eq{
					pk: id,
				}).
				PlaceholderFormat(sq.Dollar)

			rows, err := b.RunWith(s.db).Query()
			defer rows.Close()
			if err != nil {
				return nil, err
			}

			var key string
			var value interface{}

			for rows.Next() {
				err := rows.Scan(&key, &value)
				if err != nil {
					return nil, err
				}
				values[key] = value
			}

		default:
			panic("not implemented")
		}
	}

	result := map[string]interface{}{}
	for _, attrName := range attrList {
		if val, ok := values[attrName]; ok {
			result[attrName] = val
		}
	}

	return result, nil
}

func HasAttribute(list store.AttrList, name store.AttributeName) bool {
	for _, item := range list {
		if item == name {
			return true
		}
	}

	return false
}

func (s *Store) GetAttributes(name store.EntityName) store.AttrList {
	if s.shape.Primary.(store.Entity).Name == name {
		result := store.AttrList{}
		for _, attr := range s.shape.Primary.(store.Entity).Attributes {
			result = append(result, attr.(store.Attribute).Name)
		}
		return result
	}
	panic("does not exists")
}

func (s *Store) InitiateShape() error {
	//s.generateRelation(s.shape)
	sql := s.generateRelation2(s.shape)
	_, err := s.db.Exec(sql)
	return err
}

//func (s *Store) generateRelation(relation store.PrimaryWithMany) {
//	schema := &strings.Builder{}
//	gen := func(shape store.Shape, tablePrefix string) {
//		switch e := shape.(type) {
//		case store.Entity:
//			tableName := tablePrefix + e.Name
//			fmt.Fprintf(schema, "create table if not exists %s (%s_id serial);\n", tableName, tableName)
//			for _, attr := range e.Attributes {
//				switch a := attr.(type) {
//				case store.Attribute:
//					fmt.Fprintf(schema, "alter table %s add column if not exists %s_%s %s;\n", tableName, tableName, a.Name, store.MapType(a.Type, s))
//				default:
//					panic("not implemented")
//				}
//			}
//		default:
//			panic("not implemented")
//		}
//
//	}
//
//	gen(relation.Primary, "")
//	for _, entity := range relation.Secondaries {
//		gen(entity, relation.Primary.(store.Entity).Name+"_")
//	}
//
//	fmt.Println(schema.String())
//}

func (s *Store) generateRelation2(relation store.PrimaryWithMany) string {
	schema := &strings.Builder{}

	primaryKeyName := s.getPrimaryKey(relation.Primary, relation.Primary)
	gen := func(shape store.Shape, tablePrefix string) {
		switch e := shape.(type) {
		case store.Entity:
			for _, attr := range e.Attributes {
				switch a := attr.(type) {
				case store.Attribute:
					tableName := s.getColumnName(e, relation.Primary, a.Name)
					fmt.Fprintf(schema,
						"create table if not exists %s (%s uuid, key text, value text, version int default 1, unique (%s, key));\n",
						tableName, primaryKeyName, primaryKeyName)
				default:
					panic("not implemented")
				}
			}
		default:
			panic("not implemented")
		}
	}

	gen(relation.Primary, "")
	for _, entity := range relation.Secondaries {
		gen(entity, relation.Primary.(store.Entity).Name+"_")
	}

	return schema.String()
}

func (s *Store) getTableName(entity, primary store.Shape) string {
	if EntityEqual(primary, entity) {
		return "eav__" + entity.(store.Entity).Name
	}

	return "eav__" + primary.(store.Entity).Name + "__" + entity.(store.Entity).Name
}

func (s *Store) getColumnName(entity store.Entity, primary store.Shape, key store.AttributeName) string {
	for _, attr := range entity.Attributes {
		a := attr.(store.Attribute)
		if a.Name == key {
			return s.getTableName(entity, primary) + "__" + strings.ReplaceAll(store.MapType(a.Type, s).(string), " ", "_")
		}
	}

	panic(fmt.Errorf("no column with name %s in entity %#v", key, entity))
}

func (s *Store) getPrimaryKey(_, primary store.Shape) string {
	return s.getTableName(primary, primary) + "_id"
}

func EntityEqual(a, b store.Shape) bool {
	return reflect.DeepEqual(a, b)
}

func FindEntity(shape store.PrimaryWithMany, name store.EntityName) (store.Entity, bool) {
	if shape.Primary.(store.Entity).Name == name {
		return shape.Primary.(store.Entity), true
	}

	for _, secondary := range shape.Secondaries {
		if secondary.(store.Entity).Name == name {
			return secondary.(store.Entity), true
		}
	}

	return store.Entity{}, false
}

func Validate(shape store.PrimaryWithMany) error {
	// TODO implement validation
	// (1) there cannot be duplicates names of entities in whole relation
	// (2) relation is always 1:1
	// (3) names must be normilised
	//  - to lower, cammel_case and snake case must be equal, etc to prevent duplicates
	//  - underscore(_), separator(-) all allowed but only once, repeated sequences are not allowed (__ or --)
	return nil
}
