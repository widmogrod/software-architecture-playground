package elasticsearch

import "github.com/brianvoe/gofakeit/v6"

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
