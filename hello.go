package kgetset

import (
	"reflect"

	"encoding/json"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

var helloCRDInst *unstructured.Unstructured = &unstructured.Unstructured{
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

var helloInst *unstructured.Unstructured = &unstructured.Unstructured{
	Object: map[string]interface{}{
		"kind":       "Hello",
		"apiVersion": "v1",
		"metadata": map[string]interface{}{
			"name":      "my-hello",
			"namespace": "default",
			"labels": map[string]string{
				"app": "testing",
			},
		},
		"spec": map[string]interface{}{
			"message": "Hello There!!!",
		},
		"status": map[string]interface{}{
			"phase": "Up",
		},
	},
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Hello struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelloSpec   `json:"spec,omitempty"`
	Status HelloStatus `json:"status,omitempty"`
}

type HelloSpec struct {
	Message string `json:"message"`
}

type HelloStatus struct {
	Phase string `json:"phase"`
}

type crdsuite struct {
	// crd definition given to cluster
	input *unstructured.Unstructured

	// crd definition fetched from cluster
	output *unstructured.Unstructured

	client            *unclient
	resourceInterface dynamic.ResourceInterface

	abstract
}

// compile time check if crd implements Testsuite
var _ Testsuite = &crdsuite{}

// function based option that helps in building
// a crd instance
type crdOption func(*crdsuite)

func CRD(options ...crdOption) Testsuite {
	c := &crdsuite{
		input:  helloCRDInst,
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

func (c *crdsuite) getResourceInterfaceOrDie() dynamic.ResourceInterface {
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

func (c *crdsuite) setup() (err error) {
	ri := c.getResourceInterfaceOrDie()
	// create at K8s
	_, err = ri.Create(c.input, metav1.CreateOptions{})
	// fetch the same from K8s
	c.output, err = ri.Get(c.input.GetName(), metav1.GetOptions{})
	return
}

func (c *crdsuite) postsetup() error {
	if reflect.DeepEqual(c.input, c.output) {
		return nil
	}

	i, err := json.MarshalIndent(c.input, "", ".")
	if err != nil {
		return err
	}
	o, err := json.MarshalIndent(c.output, "", ".")
	if err != nil {
		return err
	}

	return errors.Errorf(
		"mismatch found:\ninput definition:--\n%s\noutput definition:--\n%s",
		string(i),
		string(o),
	)
}

func (c *crdsuite) teardown() error {
	ri := c.getResourceInterfaceOrDie()
	deletePropagation := metav1.DeletePropagationForeground
	return ri.Delete(
		c.input.GetName(),
		&metav1.DeleteOptions{PropagationPolicy: &deletePropagation},
	)
}
