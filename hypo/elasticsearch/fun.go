package elasticsearch

import (
	"github.com/brianvoe/gofakeit/v6"
	"math/rand"
)

type Fun struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Content  string   `json:"content"`
	Location Location `json:"location"`
}

type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func GenFun() Fun {
	return Fun{
		ID:   gofakeit.UUID(),
		Name: gofakeit.Name(),
		Location: Location{
			Lat: gofakeit.Latitude(),
			Lon: gofakeit.Longitude(),
		},
		Content: gofakeit.HipsterSentence(10),
	}
}

type (
	Schema struct {
		Name   string
		Fields []Field
	}
	Field struct {
		Name  string
		Types Types
	}
	Types struct {
		String bool
		Record []Field
	}
)

type GenContext struct {
	MaxFields         int
	ProbOfRecord      float64
	GenerateFieldName func() string
}

func GenSchema(c *GenContext) Schema {
	//fieldNames := []string{
	//	"question",
	//	"answer",
	//	"user",
	//	""
	//}
	if c == nil {
		c = &GenContext{
			MaxFields:    20,
			ProbOfRecord: 0.5,
			GenerateFieldName: func() string {
				return gofakeit.Verb()
			},
		}
	}
	return Schema{
		Name:   gofakeit.AnimalType(),
		Fields: GenFields(c),
	}
}

func GenFields(c *GenContext) []Field {
	var fields []Field
	for c.MaxFields > 0 {
		c.MaxFields--
		fields = append(fields, GenField(c))
	}
	return fields
}

func GenField(c *GenContext) Field {
	return Field{
		Name:  c.GenerateFieldName(),
		Types: GetTypes(c),
	}
}

func GetTypes(c *GenContext) Types {
	if rand.Float64() < c.ProbOfRecord {
		return Types{
			String: true,
		}
	}

	return Types{
		Record: GenFields(c),
	}
}
