package storage

type Storage[T any] interface {
	GetAs(id string, x *T) error
}

func RetriveID[T any](s Storage[T], id string) (T, error) {
	var x T
	err := s.GetAs(id, &x)
	if err != nil {
		return x, err
	}
	return x, nil
}

type UpdateRecords[T any] struct {
	Saving map[string]T
}
