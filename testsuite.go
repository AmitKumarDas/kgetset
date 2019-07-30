package kgetset

import (
	"fmt"
)

type Testsuite interface {
	Test() error
}

type BDD interface {
	Given() error
	When() error
	Then() error
}

// TestAbstract as the name suggests abstracts some
// of the common features required by instances
// implementing bdd or testsuite interface
type TestAbstract struct {
	Steps []func() error

	Setupfn     func() error
	Postsetupfn func() error

	Teardownfn     func() error
	Postteardownfn func() error

	Givenfn func() error
	Whenfn  func() error
	Thenfn  func() error
}

func (t *TestAbstract) Setup() error {
	if t.Setupfn == nil {
		return nil
	}
	fmt.Println("will execute setup")
	return t.Setupfn()
}

func (t *TestAbstract) Postsetup() error {
	if t.Postsetupfn == nil {
		return nil
	}
	fmt.Println("will execute postsetup")
	return t.Postsetupfn()
}

func (t *TestAbstract) Teardown() error {
	if t.Teardownfn == nil {
		return nil
	}
	fmt.Println("will execute teardown")
	return t.Teardownfn()
}

func (t *TestAbstract) Postteardown() error {
	if t.Postteardownfn == nil {
		return nil
	}
	fmt.Println("will execute postteardown")
	return t.Postteardownfn()
}

func (t *TestAbstract) Given() error {
	if t.Givenfn == nil {
		return nil
	}
	fmt.Println("will execute given")
	return t.Givenfn()
}

func (t *TestAbstract) When() error {
	if t.Whenfn == nil {
		return nil
	}
	fmt.Println("will execute when")
	return t.Whenfn()
}

func (t *TestAbstract) Then() error {
	if t.Thenfn == nil {
		return nil
	}
	fmt.Println("will execute then")
	return t.Thenfn()
}

func (t *TestAbstract) Test() error {
	var steps = t.Steps

	if len(steps) == 0 {
		steps = []func() error{
			t.Setup,
			t.Postsetup,
			t.Given,
			t.When,
			t.Then,
			t.Teardown,
			t.Postteardown,
		}
	}

	for _, fn := range steps {
		err := fn()
		if err != nil {
			// try teardown anyway
			e := t.Teardown()
			fmt.Printf("teardown was attempted for the setup: %+v", e)

			return err
		}
	}
	return nil
}
