package base

func ValOrNil[T any](p *T) any {
	if p == nil {
		return nil
	}
	return *p
}
