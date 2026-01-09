package set

import "encoding/json"

type Set[T comparable] map[T]struct{}

func New[T comparable]() Set[T] {
	return make(Set[T])
}

func NewFromSlice[T comparable](items []T) Set[T] {
	s := make(Set[T])
	for _, item := range items {
		s[item] = struct{}{}
	}
	return s
}

func NewFromMap[T comparable, V any](items map[T]V) Set[T] {
	s := make(Set[T])
	for item := range items {
		s[item] = struct{}{}
	}
	return s
}

func NewWithItems[T comparable](items ...T) Set[T] {
	s := make(Set[T])
	for _, item := range items {
		s[item] = struct{}{}
	}
	return s
}

func (s Set[T]) Copy() Set[T] {
	newSet := make(Set[T], len(s))
	for elem := range s {
		newSet[elem] = struct{}{}
	}
	return newSet
}

func (s Set[T]) Add(element T) {
	s[element] = struct{}{}
}

func (s Set[T]) Remove(element T) {
	delete(s, element)
}

func (s Set[T]) Contains(element T) bool {
	_, exists := s[element]
	return exists
}

func (s Set[T]) ContainsSome(other Set[T]) bool {
	for elem := range s {
		if other.Contains(elem) {
			return true
		}
	}
	return false
}

func (s Set[T]) ContainsAll(other Set[T]) bool {
	for elem := range other {
		if !s.Contains(elem) {
			return false
		}
	}
	return true
}

func (s Set[T]) Intersection(other Set[T]) Set[T] {
	result := New[T]()
	for elem := range s {
		if other.Contains(elem) {
			result.Add(elem)
		}
	}
	return result
}

func (s Set[T]) Difference(other Set[T]) Set[T] {
	result := New[T]()
	for elem := range s {
		if !other.Contains(elem) {
			result.Add(elem)
		}
	}
	return result
}

func (s Set[T]) Union(other Set[T]) Set[T] {
	result := New[T]()
	for elem := range s {
		result.Add(elem)
	}
	for elem := range other {
		result.Add(elem)
	}
	return result
}

func (s Set[T]) Equals(other Set[T]) bool {
	if len(s) != len(other) {
		return false
	}
	for elem := range s {
		if !other.Contains(elem) {
			return false
		}
	}
	return true
}

func (s Set[T]) ToSlice() []T {
	slice := make([]T, 0, len(s))
	for elem := range s {
		slice = append(slice, elem)
	}
	return slice
}

func (s Set[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.ToSlice())
}

func (s *Set[T]) UnmarshalJSON(data []byte) error {
	var slice []T
	if err := json.Unmarshal(data, &slice); err != nil {
		return err
	}
	*s = NewWithItems(slice...)
	return nil
}
