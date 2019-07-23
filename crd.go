package kgetset

import (
	"reflect"

	"encoding/json"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

var crdInst *unstructured.Unstructured = &unstructured.Unstructured{
	Object: map[string]interface{}{
		"kind":       "CustomResourceDefinition",
		"apiVersion": "apiextensions.k8s.io/v1beta1",
		"metadata": map[string]interface{}{
			"name": "hellos.openebs.io",
		},
		"spec": map[string]interface{}{
			"group":   "openebs.io",
			"version": "v1",
			"scope":   "Namespaced",
			"names": map[string]interface{}{
				"plural":     "hellos",
				"singular":   "hello",
				"kind":       "Hello",
				"shortNames": []string{"hello"},
			},
		},
	},
}

type crd struct {
	input  *unstructured.Unstructured
	output *unstructured.Unstructured

	client            *unclient
	resourceInterface dynamic.ResourceInterface

	abstract

	postthenfn func() error
}

// compile time check if crd implements testsuite
var _ testsuite = &crd{}

func newcrd() *crd {
	c := &crd{
		input:  crdInst,
		client: newUnClientOrDie(),
	}

	c.setupfn = func() error {
		return c.setup()
	}

	c.teardownfn = func() error {
		return c.teardown()
	}

	c.thenfn = func() error {
		return c.then()
	}

	c.postthenfn = func() error {
		return c.postthen()
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
	_, err = ri.Create(c.input, metav1.CreateOptions{})
	return
}

func (c *crd) teardown() error {
	ri := c.getResourceInterfaceOrDie()
	deletePropagation := metav1.DeletePropagationForeground
	return ri.Delete(
		c.input.GetName(),
		&metav1.DeleteOptions{PropagationPolicy: &deletePropagation},
	)
}

func (c *crd) then() (err error) {
	ri := c.getResourceInterfaceOrDie()
	c.output, err = ri.Get(c.input.GetName(), metav1.GetOptions{})
	return
}

func (c *crd) postthen() error {
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
		"mismatch found:\ninput:--\n%s\noutput:--\n%s",
		string(i),
		string(o),
	)
}

func (c *crd) test() error {
	c.steps = []func() error{
		c.setupfn,
		c.givenfn,
		c.whenfn,
		c.thenfn,
		c.postthenfn,
		c.teardownfn,
	}

	return c.abstract.test()
}
