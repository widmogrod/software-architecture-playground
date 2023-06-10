package lets_build_db

import (
	"encoding/json"
	"fmt"
	"github.com/google/btree"
	"io"
	"strings"
)

const (
	KEY = 0
	VAL = 1

	TOMBSTONE = "**TODO**TOMBSTONE**"
)

type (
	KList       = []string
	KV          = [2]string
	KVSortedSet = []KV
	AppendLog   = []KVSortedSet
	Segment     = []KVSortedSet
)

func Set(appendLog AppendLog, kvSet KVSortedSet) AppendLog {
	appendLog = append(appendLog, kvSet)
	return appendLog
}

// Get
// Currently cost of operations is O(n) where n is len of append log
// optimisation would be knowledge at which positions in append log key exists (revers index)
//
// When appendLog will be too big to fit in memory, it will be flushed on disk
// so in cases when disk-io will be necessary to return information about record
// revers index will be helpful to locate segments of append log that contain information
//
// Compaction process will also have to update revers-index information.
// Both compaction and revers index should reduce number of disk-io operations necessary to get values for keys
//
// Now question...
// - what would happen if instead of disk-io it would be network-io?

// - what would happen if instead of reverse-index lookup
//		`key -> []segmentName` and search for key in segment (smaller reverse index)
//		`key -> []{segmentName, position}` and read value directly (bigger reverse index)
//	 it would be
//		`url://key -> segment - but depending on size of segment
//					  			it could be much wise to send computation to data server rather than data
// 	    `url://key -> value   - and gets value right away, but it also means that "reverse-index"
//	    			  			lockup base on url is deterministic, system knows on which server the latest value leaves
// 								Ideas that could be use: consistent hashing
//
// Now question what is difference between such database system and
// - federated query - AWS Athena + S3, Hive + Hadoop, GraphQL + μservice

func Get(appendLog AppendLog, keys []string) (res KVSortedSet) {
	eachSegmentKV(appendLog, latestOnly(nonDeleted(func(kv KV) {
		key := kv[KEY]
		var newKeys []string
		for _, k := range keys {
			if k == key {
				res = append(res, kv)
			} else {
				newKeys = append(newKeys, k)
			}
		}

		keys = newKeys
	})))

	return res
}

// Delete works in the same manner as Set
// just a value that is set is tombstone
// that will be physically removed during compaction
// and during Get or Find operations, such values will be threaded as deleted
func Delete(s Segment, keys KList) Segment {
	kvs := make(KVSortedSet, len(keys))
	for i, key := range keys {
		kvs[i] = KV{key, TOMBSTONE}
	}
	return Set(s, kvs)
}

// Compact function aims to reduce size of append log by removing key-value pairs that are overwritten by newer values
// Because appendLog can be in use, compacting function should work on flushed & immutable segments.
// - Segments sizes are not guaranteed to be the same.
//   Mostly because they can be flushed when max-size is reach or max-time for segment to leave in memory.
//   Which means that order in which segments will be merged should improve performance and use in implementation [√]
// - When two segment don't share a key, then it needs to be decided either
//   - don't modify two segments
//   - always create new segments
//      - this option make sence when segments are uneven and created as time-snapshot
//        rather than max-segment-data,
//      - but for time service data, or append data, compaction won't bring many benefits
// 		  so compaction should be configurable
// - Segment "a' is older than segment "b" that is created later, time wise
func Compact(a, b Segment) Segment {
	c := &collect{
		kvSet:  KVSortedSet{},
		unique: map[string]struct{}{},
	}
	eachSegmentKV(b, nonDeleted(collectUnique(c)))
	eachSegmentKV(a, nonDeleted(collectUnique(c)))
	return Segment{c.kvSet}
}

type (
	collect struct {
		kvSet  KVSortedSet
		unique map[string]struct{}
	}
)

func collectUnique(c *collect) func(kv KV) {
	return func(kv KV) {
		key := kv[KEY]
		if _, ok := c.unique[key]; !ok {
			c.unique[key] = struct{}{}
			c.kvSet = append(c.kvSet, kv)
		}
	}
}

func eachKV(kvSet KVSortedSet, f func(kv KV)) {
	for i := 0; i < len(kvSet); i++ {
		f(kvSet[i])
	}
}

func eachSegmentKV(s Segment, f func(kv KV)) {
	for i := len(s) - 1; i >= 0; i-- {
		kvSet := s[i]
		eachKV(kvSet, f)
	}
}

func eachSegmentKVI(s Segment, f func(kv KV, segmentIdx int)) {
	for i := len(s) - 1; i >= 0; i-- {
		kvSet := s[i]
		for j := 0; j < len(kvSet); j++ {
			f(kvSet[j], i)
		}
	}
}

func nonDeleted(fn func(kv KV)) func(kv KV) {
	deleted := make(map[string]struct{})
	return func(kv KV) {
		key := kv[KEY]
		if _, ok := deleted[key]; ok {
			return
		}

		isDeleted := kv[VAL] == TOMBSTONE
		if isDeleted {
			deleted[key] = struct{}{}
			return
		}

		fn(kv)
	}
}

func latestOnly(fn func(kv KV)) func(kv KV) {
	latest := make(map[string]struct{})
	return func(kv KV) {
		key := kv[KEY]
		if _, ok := latest[key]; ok {
			// don't process same keys twice
			return
		}

		latest[key] = struct{}{}

		fn(kv)
	}
}

// Find return key sets that match testFn criteria.
// Interestingly, from one perspective such scanning and filtering may be suboptimal
// Some "find" operations, can have optimise indices to perform faster lookup operations
// In this case I make simplification to start moving forward
func Find(s Segment, testFn func(KV) bool, limit uint) KVSortedSet {
	var res KVSortedSet
	eachSegmentKV(s, latestOnly(nonDeleted(func(kv KV) {
		if limit == 0 {
			// TODO `eachSegment` should have limit to stop computation
			// like maybe return false, without it this early termination
			// will continue till the end of segment
			return
		}
		if testFn(kv) {
			res = append(res, kv)
			limit--
		}
	})))

	// TODO sort results?
	return res
}

type KeyPrefix string

const separator = "\x00"

func Pack(parts ...string) *KeyPrefix {
	r := KeyPrefix(strings.Join(parts, separator))
	return &r
}

func (a *KeyPrefix) Less(b btree.Item) bool {
	return string(*a) < string(*b.(*KeyPrefix))
}

func (a *KeyPrefix) String() string {
	return string(*a)
}

func (a *KeyPrefix) Pack(parts ...string) *KeyPrefix {
	result := a.String()
	if a.isEmpty() {
		return Pack(parts...)
	}

	for i := range parts {
		result += separator + parts[i]
	}
	return Pack(result)
}

func (a *KeyPrefix) Begin() *KeyPrefix {
	return Pack(a.String(), "\x00")
}

func (a *KeyPrefix) End() *KeyPrefix {
	return Pack(a.String(), "\xff")
}

func (a *KeyPrefix) Unpack() []string {
	return strings.Split(string(*a), separator)
}

func (a *KeyPrefix) isEmpty() bool {
	return a.String() == ""
}

func Unpack(packed string) *KeyPrefix {
	return Pack(strings.Split(packed, separator)...)
}

type KSegments struct {
	KeyPrefix *KeyPrefix
	Segments  []int
}

func (a *KSegments) Less(b btree.Item) bool {
	return a.KeyPrefix.Less(b.(*KSegments).KeyPrefix)
}

func Range(appendLog AppendLog, begin, end *KeyPrefix) (_ KVSortedSet) {
	bt := btree.New(2)

	// FIX below operations needs to be build as appendLong and db is build
	// this place is not good for this
	eachSegmentKVI(appendLog, func(kv KV, segmentIdx int) {
		key := &KSegments{
			KeyPrefix: Pack(kv[KEY]),
			Segments:  []int{segmentIdx},
		}

		if prev := bt.Get(key); prev != nil {
			// First segment is latest
			key.Segments = append(key.Segments, prev.(*KSegments).Segments...)
		}

		bt.ReplaceOrInsert(key)
	})

	var keys []*KSegments

	bt.AscendRange(
		&KSegments{KeyPrefix: begin},
		&KSegments{KeyPrefix: end},
		func(item btree.Item) bool {
			keys = append(keys, item.(*KSegments))
			return true
		},
	)

	var res KVSortedSet

	// read keys from segments
	for _, ks := range keys {
		segmentIdx := ks.Segments[0]

		// fetch segment
		eachKV(appendLog[segmentIdx], func(kv KV) {
			if kv[KEY] == ks.KeyPrefix.String() {
				res = append(res, kv)
			}
		})
	}

	return res
}

type (
	Key     = string
	Version = int

	Operation struct {
		Set    *SetOp    `json:"set"`
		Delete *DeleteOp `json:"delete"`
	}
	SetOp struct {
		//version Version
		KvSet KVSortedSet `json:"kvSet"`
	}
	DeleteOp struct {
		//version Version
		Keys []Key `json:"keys"`
	}
)

type (
	NaiveDBState struct {
		memory1            Segment
		keysAndSegmentsMap *btree.BTree
	}
)

func (state *NaiveDBState) Set(kvSet KVSortedSet) error {
	state.memory1 = Set(state.memory1, kvSet)
	return nil
}

func (state *NaiveDBState) Delete(keys []Key) error {
	state.memory1 = Delete(state.memory1, keys)
	return nil
}

// Restore recreate database from last snapshot
// to restore it we have to get stream of bytes, that can be feed from any source and any size
func (state *NaiveDBState) Restore(stream io.Reader) error {
	// TODO remove json encoding
	decoder := json.NewDecoder(stream)
	decoder.UseNumber()

	// Mark database as restoring

	// Read commands & apply commands
	var err error
	for {
		op := new(Operation)
		err = decoder.Decode(op)
		if err == io.EOF {
			break
		}

		if err == nil {
			if op.Delete != nil {
				err = state.Delete(op.Delete.Keys)
			} else if op.Set != nil {
				err = state.Set(op.Set.KvSet)
			} else {
				err = fmt.Errorf("restor: unknow op %#v", op)
			}
		}

		if err != nil {
			return fmt.Errorf("restore: failed: %w", err)
		}
	}

	// Because snapshot point can be old,
	// to catch up to the newest updates Restore can ask for updates from other nodes
	// this function is much simpler, and focus on restoring state
	// Mark database as ready to accept connections
	return nil
}

func Append() {

}

func Stream() {
	// reads from append log operations that are mutations
	// Set(kvSet)
	// Delete(kvSet)
}
