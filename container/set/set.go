package set

type Set map[interface{}]int

func New(values ...interface{}) Set {
	s := Set{}
	if len(values) > 0 {
		for _, value := range values {
			s.Add(value)
		}
	}
	return s
}

func (s Set) Has(value interface{}) (exist bool) {
	if value == nil || s == nil {
		return false
	}

	_, exist = s[value]
	return
}

func (s Set) Add(value interface{}) Set {
	if value == nil || s == nil {
		return s
	}

	s[value] = 1
	return s
}

func (s Set) Remove(value interface{}) Set {
	if _, exist := s[value]; exist {
		delete(s, value)
	}
	return s
}

func (s Set) Union(values ...interface{}) Set {
	if s != nil && len(values) > 0 {
		for _, value := range values {
			s.Add(value)
		}
	}
	return s
}

func (s Set) ForEach(fun func(value interface{}) error) error {
	for k, _ := range s {
		err := fun(k)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s Set) Len() int {
	if s == nil {
		return 0
	}
	return len(s)
}
