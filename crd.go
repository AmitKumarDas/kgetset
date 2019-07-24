package kgetset

import (
	"reflect"

	"encoding/json"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

type crd struct {
	// crd definition given to cluster
	input *unstructured.Unstructured

	// crd definition fetched from cluster
	output *unstructured.Unstructured

	client            *unclient
	resourceInterface dynamic.ResourceInterface

	abstract
}

// compile time check if crd implements testsuite
var _ testsuite = &crd{}

// function based option that helps in building
// a crd instance
type crdOption func(*crd)

func newCRD(options ...crdOption) *crd {
	c := &crd{
		input:  HelloCRD,
		client: newUnClientOrDie(),
	}

	c.setupfn = func() error {
		return c.setup()
	}

	c.postsetupfn = func() error {
		return c.postsetup()
	}

	c.teardownfn = func() error {
		return c.teardown()
	}

	for _, o := range options {
		o(c)
	}

	return c
}

func (c *crd) getResourceInterfaceOrDie() dynamic.ResourceInterface {
	if c.resourceInterface != nil {
		return c.resourceInterface
	}

	ri, err := c.client.getResourceInterface(
		c.input.GroupVersionKind(),
		c.input.GetNamespace(),
	)
	if err != nil {
		panic(err)
	}

	c.resourceInterface = ri
	return ri
}

func (c *crd) setup() (err error) {
	ri := c.getResourceInterfaceOrDie()
	// create at K8s
	_, err = ri.Create(c.input, metav1.CreateOptions{})
	// fetch the same from K8s
	c.output, err = ri.Get(c.input.GetName(), metav1.GetOptions{})
	return
}

func (c *crd) postsetup() error {
	icopy := c.input.DeepCopyObject()
	ocopy := c.output.DeepCopyObject()

	if reflect.DeepEqual(icopy, ocopy) {
		return nil
	}

	i, err := json.MarshalIndent(icopy, "", ".")
	if err != nil {
		return err
	}
	o, err := json.MarshalIndent(ocopy, "", ".")
	if err != nil {
		return err
	}

	return errors.Errorf(
		"mismatch found:\ninput definition:--\n%s\noutput definition:--\n%s",
		string(i),
		string(o),
	)
}

func (c *crd) teardown() error {
	ri := c.getResourceInterfaceOrDie()
	deletePropagation := metav1.DeletePropagationForeground
	return ri.Delete(
		c.input.GetName(),
		&metav1.DeleteOptions{PropagationPolicy: &deletePropagation},
	)
}
