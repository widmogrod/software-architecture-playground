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

	d.appendLog = Set(d.appendLog, Flatten(doc, Pack("$M", "D", "$M", doc["$docId"].(string))))
	return doc, nil
}

func (d *DocDB) Get(docId string) (_ MapAny, _ error) {
	keyPrefix := Pack("$M", "D", "$M", docId)
	kvSet := Range(d.appendLog, keyPrefix.Begin(), keyPrefix.End())
	result := Unflatten(kvSet).(MapAny)
	return result["D"].(MapAny)[docId].(MapAny), nil
}

func Flatten(in interface{}, prefix *KeyPrefix) (_ KVSortedSet) {
	var result KVSortedSet
	return flatten(prefix, in, result)
}

func flatten(prefix *KeyPrefix, in interface{}, result KVSortedSet) KVSortedSet {
	switch x := in.(type) {
	case MapAny:
		for k := range x {
			v := x[k]
			result = flatten(prefix.Pack("$M", k), v, result)
		}
	case ListAny:
		for k := range x {
			v := x[k]
			result = flatten(prefix.Pack("$L", strconv.Itoa(k)), v, result)
		}
	case string:
		result = append(result, KV{prefix.Pack("$S").String(), x})
	case int:
		result = append(result, KV{prefix.Pack("$I").String(), strconv.Itoa(x)})
	case float64:
		result = append(result, KV{prefix.Pack("$F").String(), fmt.Sprintf("%f", x)})
	case bool:
		if x {
			result = append(result, KV{prefix.Pack("$B").String(), "y"})
		} else {
			result = append(result, KV{prefix.Pack("$B").String(), "n"})
		}
	default:
		panic(fmt.Errorf("unsupported type %v", x))
	}

	return result
}

func Unflatten(kvSet KVSortedSet) (_ interface{}) {
	var doc interface{}
	eachKV(kvSet, func(kv KV) {
		parts := Unpack(kv[KEY]).Unpack()
		doc = unflatten(parts, 0, doc, kv[VAL])
	})
	return doc
}

func unflatten(parts []string, index int, prev interface{}, val string) (_ interface{}) {
	switch parts[index] {
	case "$M":
		switch prevT := prev.(type) {
		case MapAny:
			key := parts[index+1]
			next, _ := prevT[key]
			prev.(MapAny)[key] = unflatten(parts, index+2, next, val)
			return prev

		case *ListAny:
			key := parts[index+1]
			i, _ := strconv.Atoi(key)
			var next interface{}
			if len(*prevT) > i {
				next = (*prevT)[i]
				(*prevT)[i] = unflatten(parts, index+2, next, val)
			} else {
				*prevT = append(*prevT, unflatten(parts, index+2, next, val))
			}
			return prev

		default:
			next := make(MapAny)
			return unflatten(parts, index, next, val)
		}

	case "$L":
		switch prevT := prev.(type) {
		case MapAny:
			key := parts[index+1]
			next, _ := prevT[key]
			prev.(MapAny)[key] = unflatten(parts, index+2, next, val)
			return prev

		case *ListAny:
			key := parts[index+1]
			i, _ := strconv.Atoi(key)
			var next interface{}
			if len(*prevT) > i {
				next = (*prevT)[i]
				(*prevT)[i] = unflatten(parts, index+2, next, val)
			} else {
				*prevT = append(*prevT, unflatten(parts, index+2, next, val))
			}

			return prev
		default:
			next := new(ListAny)
			return unflatten(parts, index, next, val)
		}

	// Leaves
	case "$S":
		return val
	case "$I":
		v, _ := strconv.Atoi(val)
		return v
	case "$F":
		v, _ := strconv.ParseFloat(val, 64)
		return v
	case "$B":
		return val == "y"
	}

	panic(fmt.Errorf("you should NEVER reach this place. \n"+
		"parts = %v\n"+
		"index = %v\n"+
		"prev = %v\n"+
		"val = %v\n", parts, index, prev, val))
}
