package postgresql

import (
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/widmogrod/software-architecture-playground/datasim/essence/algebra/store"
	"strings"
)

var (
	_ store.TypeVisitor = &ToPostgreSQL{}
)

type ToPostgreSQL struct{}

func (i *ToPostgreSQL) VisitInt(x store.Int) interface{} {
	return "int"
}

func (i *ToPostgreSQL) VisitString(x store.String) interface{} {
	return "text"
}

func (i *ToPostgreSQL) VisitDateTime(x store.DateTime) interface{} {
	return "timestamp with time zone"
}

var (
	_ store.RelationVisitor = &GenerateEAV{}
	_ store.ShapeVisitor    = &GenerateEAV{}
)

type GenerateEAV struct {
	schema  *strings.Builder
	shape store.Relation
	//primary store.Shape
}

func (i *GenerateEAV) VisitEntity(x store.Entity) interface{} {
	primaryKeyName := i.getPrimaryKey(i.primary, i.primary)
	for _, attr := range x.Attributes {
		switch a := attr.(type) {
		case store.Attribute:
			tableName := i.getColumnName(x, i.primary, a.Name)
			fmt.Fprintf(i.schema,
				"create table if not exists %s (%s uuid, key text, value text, version int default 1, unique (%s, key));\n",
				tableName, primaryKeyName, primaryKeyName)
		default:
			panic("not implemented")
		}
	}
}

func (i *GenerateEAV) VisitPrimaryWithMany(relation store.PrimaryWithMany) interface{} {
	i.primary = relation.Primary
	store.MapShape(relation.Primary, i)
	for _, entity := range relation.Secondaries {
		store.MapShape(entity, i)
	}

	return i.schema.String()
}

func (s *GenerateEAV) getTableName(entity, primary store.Shape) string {
	if EntityEqual(primary, entity) {
		return "eav__" + entity.(store.Entity).Name
	}

	return "eav__" + primary.(store.Entity).Name + "__" + entity.(store.Entity).Name
}

func (s *GenerateEAV) getColumnName(entity store.Entity, primary store.Shape, key store.AttributeName) string {
	for _, attr := range entity.Attributes {
		a := attr.(store.Attribute)
		if a.Name == key {
			return s.getTableName(entity, primary) + "__" + strings.ReplaceAll(store.MapType(a.Type, &ToPostgreSQL{}).(string), " ", "_")
		}
	}

	panic(fmt.Errorf("no column with name %s in entity %#v", key, entity))
}

func (s *GenerateEAV) getPrimaryKey(_, primary store.Shape) string {
	return s.getTableName(primary, primary) + "_id"
}

var (
	_ store.Storer = &Store{}
)

type Interpretation struct {
	db *sql.DB
}

func (s *Interpretation) VisitInitiate(x store.Initiate) interface{} {
	builder := &GenerateEAV{
		shape:  x.Relation,
	}
	_ = store.MapRelation(x.Relation, builder)
	_, err := s.db.Exec(builder.schema.String())
	return err
}

func (s *Interpretation) VisitSetPrimaryAttr(x store.SetPrimaryAttr) interface{} {
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

func NewStore2(db *sql.DB, relation store.PrimaryWithMany) *Store2 {
	return &Store2{
		db:    db,
		shape: relation,
	}
}

var _ store.Storer = &Store2{}

type Store2 struct {
	db    *sql.DB
	shape store.Relation
}

func (s Store2) Set(id store.EntityID, name store.EntityName, key store.AttributeName, value store.AttributeValue) error {
	panic("implement me")
}

func (s Store2) Get(id store.EntityID, name store.EntityName, strings store.AttrList) (interface{}, error) {
	panic("implement me")
}

func (s Store2) GetAttributes(name store.EntityName) store.AttrList {
	panic("implement me")
}

func (s *Store2) InitiateShape() error {
	i := &Interpretation{db: s.db}
	err := store.MapOperations(store.Initiate{Relation: s.shape}, i)
	if err2, ok := err.(error); ok && err2 != nil {
		return err2
	}

	return nil
}
