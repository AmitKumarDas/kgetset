package hello

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
  k8s "github.com/AmitKumarDas/kgetset"
  "github.com/AmitKumarDas/kgetset/unstruct"
)

type TestA struct {
	// crd definition given to cluster
	input *unstructured.Unstructured

	// crd definition fetched from cluster
	output *unstructured.Unstructured

	client            *k8s.Unclient
	resourceInterface dynamic.ResourceInterface

	k8s.TestAbstract
}

// compile time check if TestA implements Testsuite
var _ k8s.Testsuite = &TestA{}

func NewTestA(options ...func(*TestA)) *TestA {
	c := &TestA{
		input:  crdInst,
		client: k8s.NewUnClientOrDie(),
	}

	c.Setupfn = func() error {
		return c.setup()
	}

	c.Postsetupfn = func() error {
		return c.postsetup()
	}

	c.Teardownfn = func() error {
		return c.teardown()
	}

	for _, o := range options {
		o(c)
	}

	return c
}

func (c *TestA) getResourceInterfaceOrDie() dynamic.ResourceInterface {
	if c.resourceInterface != nil {
		return c.resourceInterface
	}

	ri, err := c.client.GetResourceInterface(
		c.input.GroupVersionKind(),
		c.input.GetNamespace(),
	)
	if err != nil {
		panic(err)
	}

	c.resourceInterface = ri
	return ri
}

func (c *TestA) setup() (err error) {
	ri := c.getResourceInterfaceOrDie()

	// create at K8s
	_, err = ri.Create(c.input, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	// fetch the same from K8s
	c.output, err = ri.Get(c.input.GetName(), metav1.GetOptions{})
	return
}

func (c *TestA) postsetup() error {
	var paths = []string{
   "spec.version",
   "spec.group",
   "spec.scope",
  }
	
	changed, err := unstruct.IsChangeStr(c.input, c.output, "metadata.name", paths...)
	if err != nil {
		return err
	}

	if !changed {
		return nil
	}

	return errors.Errorf(
		"mismatch found:\ninput definition:--\n%+v\noutput definition:--\n%+v",
		c.input,
		c.output,
	)
}

func (c *TestA) teardown() error {
	ri := c.getResourceInterfaceOrDie()
	deletePropagation := metav1.DeletePropagationForeground
	return ri.Delete(
		c.input.GetName(),
		&metav1.DeleteOptions{PropagationPolicy: &deletePropagation},
	)
}
