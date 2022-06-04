package lets_build_db

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
	found := map[string]struct{}{}
	eachSegmentKV(appendLog, nonDeleted(func(kv KV) {
		key := kv[KEY]
		if _, ok := found[key]; ok {
			return
		}

		var newKeys []string
		for _, k := range keys {
			if k == key {
				res = append(res, kv)
				found[key] = struct{}{}
			} else {
				newKeys = append(newKeys, k)
			}
		}

		keys = newKeys
	}))

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

func eachSegmentKV(s Segment, f func(kv KV)) {
	for i := len(s) - 1; i >= 0; i-- {
		kvSet := s[i]
		for j := 0; j < len(kvSet); j++ {
			f(kvSet[j])
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

// Find return key sets that match testFn criteria.
// Interestingly, from one perspective such scanning and filtering may be suboptimal
// Some "find" operations, can have optimise indices to perform faster lookup operations
// In this case I make simplification to start moving forward
func Find(s Segment, testFn func(KV) bool, limit uint) KVSortedSet {
	var res KVSortedSet
	eachSegmentKV(s, nonDeleted(func(kv KV) {
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
	}))

	// TODO sort results?
	return res
}
