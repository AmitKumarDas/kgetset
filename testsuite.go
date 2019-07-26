package kgetset

type Testsuite interface {
	Test() error
}

type BDD interface {
	Given() error
	When() error
	Then() error
}

// abstract as the name suggests abstracts some
// of the common features required by instances
// implementing bdd or testsuite interface
type abstract struct {
	steps []func() error

	setupfn     func() error
	postsetupfn func() error

	teardownfn     func() error
	postteardownfn func() error

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

func (t *abstract) postsetup() error {
	if t.postsetupfn == nil {
		return nil
	}
	return t.postsetupfn()
}

func (t *abstract) teardown() error {
	if t.teardownfn == nil {
		return nil
	}
	return t.teardownfn()
}

func (t *abstract) postteardown() error {
	if t.postteardownfn == nil {
		return nil
	}
	return t.postteardownfn()
}

func (t *abstract) Given() error {
	if t.givenfn == nil {
		return nil
	}
	return t.givenfn()
}

func (t *abstract) When() error {
	if t.whenfn == nil {
		return nil
	}
	return t.whenfn()
}

func (t *abstract) Then() error {
	if t.thenfn == nil {
		return nil
	}
	return t.thenfn()
}

func (t *abstract) Test() error {
	var steps = t.steps

	if len(t.steps) == 0 {
		steps = []func() error{
			t.setupfn,
			t.postsetupfn,
			t.givenfn,
			t.whenfn,
			t.thenfn,
			t.teardownfn,
			t.postteardownfn,
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
