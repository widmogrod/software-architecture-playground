package lets_build_db

import (
	"fmt"
	"github.com/google/uuid"
	"strconv"
)

func NewDocDb() *DocDB {
	return &DocDB{
		nextRowId: func() string {
			return uuid.New().String()
		},
	}
}

type (
	MapAny  = map[string]interface{}
	ListAny = []interface{}
)

type DocDB struct {
	appendLog AppendLog
	nextRowId func() string
}

func (d *DocDB) Save(doc MapAny) (_ MapAny, _ error) {
	if docId, ok := doc["$docId"]; ok {
		if _, ok := docId.(string); !ok {
			return nil, fmt.Errorf("$docId %s != string", docId)
		}
	} else {
		doc["$docId"] = d.nextRowId()
	}

	d.appendLog = Set(d.appendLog, Flatten(doc, Pack("D", doc["$docId"].(string))))
	return doc, nil
}

func (d *DocDB) Get(docId string) (_ MapAny, _ error) {
	keyPrefix := Pack("D", docId)
	kvSet := Range(d.appendLog, keyPrefix.Begin(), keyPrefix.End())
	result := Unflatten(kvSet)
	return result["D"].(MapAny)[docId].(MapAny), nil
}

func Flatten(in interface{}, prefix *KeyPrefix) (_ KVSortedSet) {
	var result KVSortedSet

	switch x := in.(type) {
	case MapAny:
		for k := range x {
			v := x[k]
			result = funcName(prefix.Pack(k), v, result)
		}
	case ListAny:
		for k := range x {
			v := x[k]
			result = funcName(prefix.Pack(strconv.Itoa(k)), v, result)
		}

	default:
		panic(fmt.Errorf("%v, is not iterable", in))
	}

	return result
}

func funcName(prefix *KeyPrefix, v interface{}, result KVSortedSet) KVSortedSet {
	switch v.(type) {
	case MapAny, ListAny:
		result = append(result, Flatten(v, prefix)...)
	default:
		// TODO encode type of value
		result = append(result, KV{prefix.String(), fmt.Sprintf("%s", v)})
	}
	return result
}

func Unflatten(kvSet KVSortedSet) (_ MapAny) {
	result := make(MapAny)
	eachKV(kvSet, func(kv KV) {
		var doc interface{}
		doc = result

		parts := Unpack(kv[KEY]).Unpack()
		l := len(parts)
		for i := 0; i < l; i++ {
			key := parts[i]
			isLast := l == i+1

			switch d := doc.(type) {
			case MapAny:
				if isLast {
					d[key] = kv[VAL]
					return
				}
				if _, ok := d[key]; !ok {
					next := parts[i+1]
					// FIX: this type of heuristic don't allow map keys to be numeric! add value types
					_, err := strconv.Atoi(next)
					if err != nil {
						d[key] = make(MapAny)
					} else {
						d[key] = new(ListAny)
					}
				}
				doc = d[key]
			case *ListAny:
				if isLast {
					*d = append(*d, kv[VAL])
					return
				}

				next := parts[i+1]
				// FIX: this type of heuristic don't allow map keys to be numeric! add value types
				idx, err := strconv.Atoi(next)
				if err != nil {
					// FIX: this type of heuristic don't allow map keys to be numeric! add value types
					keyIdx, _ := strconv.Atoi(key)
					if len(*d) <= keyIdx {
						doc = make(MapAny)
						*d = append(*d, doc)
					} else {
						// Retrieve previously created map
						// That is stored under given index
						doc = (*d)[keyIdx]
					}
				} else {
					if idx == 0 {
						// Create list
						doc = new(ListAny)
						*d = append(*d, doc)
					} else {
						// Retrieve previously created list
						// and append to it
						doc = (*d)[len(*d)-1]
					}
				}
			default:
				panic(fmt.Errorf("%#v is unsupported", doc))
			}

		}
	})

	return result
}
