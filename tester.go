package kgetset

type testsuite interface {
	setup() error
	teardown() error
	test() error
}

type bdd interface {
	given() error
	when() error
	then() error
}

type abstract struct {
	steps []func() error

	setupfn    func() error
	teardownfn func() error

	givenfn func() error
	whenfn  func() error
	thenfn  func() error
}

func (t *abstract) setup() error {
	if t.setupfn == nil {
		return nil
	}
	return t.setupfn()
}

func (t *abstract) teardown() error {
	if t.teardownfn == nil {
		return nil
	}
	return t.teardownfn()
}

func (t *abstract) given() error {
	if t.givenfn == nil {
		return nil
	}
	return t.givenfn()
}

func (t *abstract) when() error {
	if t.whenfn == nil {
		return nil
	}
	return t.whenfn()
}

func (t *abstract) then() error {
	if t.thenfn == nil {
		return nil
	}
	return t.thenfn()
}

func (t *abstract) test() error {
	var steps = t.steps

	if len(t.steps) == 0 {
		steps = []func() error{
			t.setupfn,
			t.givenfn,
			t.whenfn,
			t.thenfn,
			t.teardownfn,
		}
	}

	for _, fn := range steps {
		err := fn()
		if err != nil {
			// try teardown anyway
			t.teardownfn()
			return err
		}
	}
	return nil
}
