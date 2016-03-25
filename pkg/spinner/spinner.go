package spinner

type Spinner struct {
	i int8
}

func New() *Spinner {
	return &Spinner{}
}

func (s *Spinner) Next() byte {
	defer func() {
		s.i = (s.i + 1) % 4
	}()

	switch s.i {
	case 1:
		return '/'
	case 2:
		return '-'
	case 3:
		return '\\'
	default:
		return '|'
	}
}
