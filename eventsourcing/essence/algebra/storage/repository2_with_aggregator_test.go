package storage

type User struct {
	ID   string
	Name string
	Age  int
}

//func TestNewRepositoryInMemory(t *testing.T) {
//	r := NewRepositoryWithIndexer[User]()
//	err := r.UpdateRecords(UpdateRecords[Record[User]]{
//		Saving: map[string]Record[User]{
//			"1": {
//				SessionID: "1",
//			},
//		},
//	})
//
//	assert.NoError(t, err)
//}
