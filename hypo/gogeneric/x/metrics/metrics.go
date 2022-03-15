// Package metrics provides a general mechanism for accumulating
// metrics of different values.
// Source: https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md
package metrics

import "sync"

// Metric1 accumulates metrics of a single value.
type Metric1[T comparable] struct {
	mu sync.Mutex
	m  map[T]int
}

// Add adds an instance of a value.
func (m *Metric1[T]) Add(v T) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.m == nil {
		m.m = make(map[T]int)
	}
	m.m[v]++
}

// key2 is an internal type used by Metric2.
type key2[T1, T2 comparable] struct {
	f1 T1
	f2 T2
}

// Metric2 accumulates metrics of pairs of values.
type Metric2[T1, T2 comparable] struct {
	mu sync.Mutex
	m  map[key2[T1, T2]]int
}

// Add adds an instance of a value pair.
func (m *Metric2[T1, T2]) Add(v1 T1, v2 T2) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.m == nil {
		m.m = make(map[key2[T1, T2]]int)
	}
	m.m[key2[T1, T2]{v1, v2}]++
}

// key3 is an internal type used by Metric3.
type key3[T1, T2, T3 comparable] struct {
	f1 T1
	f2 T2
	f3 T3
}

// Metric3 accumulates metrics of triples of values.
type Metric3[T1, T2, T3 comparable] struct {
	mu sync.Mutex
	m  map[key3[T1, T2, T3]]int
}

// Add adds an instance of a value triplet.
func (m *Metric3[T1, T2, T3]) Add(v1 T1, v2 T2, v3 T3) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.m == nil {
		m.m = make(map[key3[T1, T2, T3]]int)
	}
	m.m[key3[T1, T2, T3]{v1, v2, v3}]++
}

// Repeat for the maximum number of permitted arguments.
