package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"strings"
)

// HashKey is the key used in Hash maps
type HashKey struct {
	Type  ObjectType
	Value uint64
}

// Hashable is an interface for objects that can be used as hash keys
type Hashable interface {
	HashKey() HashKey
}

// HashPair represents a key-value pair in a hash
type HashPair struct {
	Key   Object
	Value Object
}

// Hash represents a hash map object
type Hash struct {
	Pairs map[HashKey]HashPair
}

// Type returns the type of the object
func (h *Hash) Type() ObjectType { return HASH_OBJ }

// Inspect returns a string representation of the hash
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

// HashKey returns a hash key for the object
func (b *Boolean) HashKey() HashKey {
	var value uint64

	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

// HashKey returns a hash key for the object
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

// HashKey returns a hash key for the object
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))

	return HashKey{Type: s.Type(), Value: h.Sum64()}
}
