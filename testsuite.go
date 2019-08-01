package kgetset

import (
	"fmt"
	"time"
)

type Testsuite interface {
	Test() error
}

type BDD interface {
	Given() error
	When() error
	Then() error
}

type TestFns []func() error

func (f TestFns) Run() error {
	for _, fn := range f {
		if fn == nil {
			continue
		}
		err := fn()
		if err != nil {
			return err
		}
	}
	return nil
}

const (
	StepSetup = iota + 1
	StepPostSetup

	StepGiven
	StepWhen
	StepThen

	StepTeardown
	StepPostTeardown
)

// TestAbstract as the name suggests abstracts some
// of the common features required by instances
// implementing bdd or testsuite interface
type TestAbstract struct {
	WaitPostSteps []int
	WaitTime      time.Duration

	Steps   []func() error
	stepIdx int

	Setupfn     func() error
	PostSetupfn func() error

	Teardownfn     func() error
	PostTeardownfn func() error

	Givenfn func() error
	Whenfn  func() error
	Thenfn  func() error
}

func (t *TestAbstract) waitPostStep() {
	if len(t.WaitPostSteps) == 0 {
		return
	}
	for idx := range t.WaitPostSteps {
		if idx == t.stepIdx {
			time.Sleep(t.WaitTime)
		}
	}
}

func (t *TestAbstract) Setup() error {
	if t.Setupfn == nil {
		return nil
	}
	fmt.Printf("[%d] executing setup\n", t.stepIdx)
	return t.Setupfn()
}

func (t *TestAbstract) PostSetup() error {
	if t.PostSetupfn == nil {
		return nil
	}
	fmt.Printf("[%d] executing postsetup\n", t.stepIdx)
	return t.PostSetupfn()
}

func (t *TestAbstract) Teardown() error {
	if t.Teardownfn == nil {
		return nil
	}
	fmt.Printf("[%d] executing teardown\n", t.stepIdx)
	return t.Teardownfn()
}

func (t *TestAbstract) PostTeardown() error {
	if t.PostTeardownfn == nil {
		return nil
	}
	fmt.Printf("[%d] executing postteardown\n", t.stepIdx)
	return t.PostTeardownfn()
}

func (t *TestAbstract) Given() error {
	if t.Givenfn == nil {
		return nil
	}
	fmt.Printf("[%d] executing given\n", t.stepIdx)
	return t.Givenfn()
}

func (t *TestAbstract) When() error {
	if t.Whenfn == nil {
		return nil
	}
	fmt.Printf("[%d] executing when\n", t.stepIdx)
	return t.Whenfn()
}

func (t *TestAbstract) Then() error {
	if t.Thenfn == nil {
		return nil
	}
	fmt.Printf("[%d] executing then\n", t.stepIdx)
	return t.Thenfn()
}

func (t *TestAbstract) Test() error {
	var steps = t.Steps

	if len(steps) == 0 {
		steps = []func() error{
			t.Setup,
			t.PostSetup,
			t.Given,
			t.When,
			t.Then,
			t.Teardown,
			t.PostTeardown,
		}
	}

	for _, fn := range steps {
		t.stepIdx++
		err := fn()
		if err != nil {
			// if error try teardown before aborting
			t.stepIdx++
			e := t.Teardown()
			fmt.Printf("testsuite teardown was attempted: %+v", e)

			return err
		}
		t.waitPostStep()
	}
	return nil
}
