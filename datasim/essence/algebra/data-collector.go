package algebra

import (
	"encoding/csv"
	"os"
)

type (
	DataCollector struct {
		entries   []entry
		//counters  map[string]int
		namespace string
	}

	entry struct {
		key    Key
		val    Value
		labels Metadata
	}

	Key      = string
	Value    = string
	Metadata = map[string]string
)

func NewDataCollector() *DataCollector {
	return &DataCollector{
		entries:   make([]entry, 0),
		//counters:  make(map[string]int),
		namespace: "default",
	}
}

func (dc *DataCollector) Push(key Key, value Value, metadata Metadata) {
	dc.entries = append(dc.entries, entry{
		key:    key,
		val:    value,
		labels: metadata,
	})
}

//func (dc *DataCollector) Increment(metadata Metadata) {
//	for key, value := range metadata {
//		dc.counters[key] += 1
//		dc.counters[key+"["+value+"]"] += 1
//	}
//}

func (dc *DataCollector) FlushCSV(path string) error {
	f, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	w := csv.NewWriter(f)
	_ = w.Write([]string{"namespace", "key", "value", "label_key", "label_value"})

	for _, e := range dc.entries {
		for lkey, lval := range e.labels {
			_ = w.Write([]string{dc.namespace, e.key, e.val, lkey, lval})
		}
	}

	//for key, value := range dc.counters {
	//	_ = w.Write([]string{dc.namespace, key, strconv.Itoa(value)})
	//}

	w.Flush()
	return w.Error()
}
