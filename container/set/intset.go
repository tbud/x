package set

type IntSet map[int]bool

func NewIntSet(items ...int) IntSet {
	s := IntSet{}
	if len(items) > 0 {
		for _, item := range items {
			s.Add(item)
		}
	}
	return s
}

func (s IntSet) Add(item interface{}) IntSet {
	if s == nil {
		return s
	}

	it, _ := item.(int)

	s[it] = true
	return s
}
