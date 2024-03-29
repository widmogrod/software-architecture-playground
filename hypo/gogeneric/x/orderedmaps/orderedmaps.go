// Package orderedmaps provides an ordered map, implemented as a binary tree.
// Source: https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md
package orderedmaps

import "github.com/widmogrod/software-architecture-playground/hypo/gogeneric/x/chans"

// Map is an ordered map.
type Map[K, V any] struct {
	root    *node[K, V]
	compare func(K, K) int
}

// node is the type of a node in the binary tree.
type node[K, V any] struct {
	k           K
	v           V
	left, right *node[K, V]
}

// New returns a new map.
// Since the type parameter V is only used for the result,
// type inference does not work, and calls to New must always
// pass explicit type arguments.
func New[K, V any](compare func(K, K) int) *Map[K, V] {
	return &Map[K, V]{compare: compare}
}

// find looks up k in the map, and returns either a pointer
// to the node holding k, or a pointer to the location where
// such a node would go.
func (m *Map[K, V]) find(k K) **node[K, V] {
	pn := &m.root
	for *pn != nil {
		switch cmp := m.compare(k, (*pn).k); {
		case cmp < 0:
			pn = &(*pn).left
		case cmp > 0:
			pn = &(*pn).right
		default:
			return pn
		}
	}
	return pn
}

// Insert inserts a new key/value into the map.
// If the key is already present, the value is replaced.
// Reports whether this is a new key.
func (m *Map[K, V]) Insert(k K, v V) bool {
	pn := m.find(k)
	if *pn != nil {
		(*pn).v = v
		return false
	}
	*pn = &node[K, V]{k: k, v: v}
	return true
}

// Find returns the value associated with a key, or zero if not present.
// The bool result reports whether the key was found.
func (m *Map[K, V]) Find(k K) (V, bool) {
	pn := m.find(k)
	if *pn == nil {
		var zero V // see the discussion of zero values, above
		return zero, false
	}
	return (*pn).v, true
}

// keyValue is a pair of key and value used when iterating.
type keyValue[K, V any] struct {
	k K
	v V
}

// InOrder returns an iterator that does an in-order traversal of the map.
func (m *Map[K, V]) InOrder() *Iterator[K, V] {
	type kv = keyValue[K, V] // convenient shorthand
	sender, receiver := chans.Ranger[kv]()
	var f func(*node[K, V]) bool
	f = func(n *node[K, V]) bool {
		if n == nil {
			return true
		}
		// Stop sending values if sender.Send returns false,
		// meaning that nothing is listening at the receiver end.
		return f(n.left) &&
			sender.Send(kv{n.k, n.v}) &&
			f(n.right)
	}
	go func() {
		f(m.root)
		sender.Close()
	}()
	return &Iterator[K, V]{receiver}
}

// Iterator is used to iterate over the map.
type Iterator[K, V any] struct {
	r *chans.Receiver[keyValue[K, V]]
}

// Next returns the next key and value pair. The bool result reports
// whether the values are valid. If the values are not valid, we have
// reached the end.
func (it *Iterator[K, V]) Next() (K, V, bool) {
	kv, ok := it.r.Next()
	return kv.k, kv.v, ok
}
