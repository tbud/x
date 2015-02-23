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

func (s IntSet) Add(item int) IntSet {
	if s == nil {
		return s
	}

	s[item] = true
	return s
}
