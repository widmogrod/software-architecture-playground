package lets_build_db

func Set(appendLog [][][2]string, set [][2]string) {
	appendLog = append(appendLog, set)
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
//

func Get(appendLog [][][2]string, keys []string) (res [][2]string) {
	found := map[string]struct{}{}
	for i := len(appendLog) - 1; i >= 0; i-- {
		set := appendLog[i]

		for is := 0; is < len(set); is++ {
			key := set[is][0]
			if _, ok := found[key]; ok {
				continue
			}

			var newKeys []string
			for _, k := range keys {
				if k == key {
					res = append(res, set[is])
					found[key] = struct{}{}
				} else {
					newKeys = append(newKeys, k)
				}
			}

			keys = newKeys
		}
	}

	return res
}
