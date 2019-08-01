package onegvkdiffschemas

import (
	"reflect"

	kgs "github.com/AmitKumarDas/kgetset"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
)

type TestA struct {
	client *kgs.DynClient

	crdGVK          schema.GroupVersionKind
	crdDynInterface dynamic.ResourceInterface

	resGVK          schema.GroupVersionKind
	resNamespace    string
	resDynInterface dynamic.ResourceInterface

	crd       *unstructured.Unstructured
	resourceA *unstructured.Unstructured
	resourceB *unstructured.Unstructured

	kgs.TestAbstract
}

// compile time check if TestA implements Testsuite
var _ kgs.Testsuite = &TestA{}

func NewTestA(options ...func(*TestA)) *TestA {
	c := &TestA{
		crd:    crdInst,
		crdGVK: crdInst.GetObjectKind().GroupVersionKind(),

		resGVK:       resourceInstA.GetObjectKind().GroupVersionKind(),
		resNamespace: resourceInstA.GetNamespace(),
		resourceA:    resourceInstA,
		resourceB:    resourceInstB,

		client: kgs.NewDynClientOrDie(),
	}

	c.Setupfn = c.createCRD
	c.PostSetupfn = func() error {
		fns := kgs.TestFns{
			c.registerScheme,
			c.refresh,
		}
		return fns.Run()
	}

	c.Whenfn = func() error {
		fns := kgs.TestFns{
			c.createA,
			c.createB,
		}
		return fns.Run()
	}

	c.Thenfn = func() error {
		fns := kgs.TestFns{
			c.getAndMatchA,
			c.getAndMatchB,
		}
		return fns.Run()
	}

	c.Teardownfn = c.deleteCRD
	c.PostTeardownfn = c.verifyNoResInstances

	for _, o := range options {
		o(c)
	}

	return c
}

func (c *TestA) getDynamicInterfaceForCRD() (dynamic.ResourceInterface, error) {
	if c.crdDynInterface != nil {
		return c.crdDynInterface, nil
	}
	cdi, err := c.client.GetResourceInterface(c.crdGVK)
	if err != nil {
		return nil, err
	}
	c.crdDynInterface = cdi
	return cdi, nil
}

func (c *TestA) getDynamicInterfaceForRes() (dynamic.ResourceInterface, error) {
	if c.resDynInterface != nil {
		return c.resDynInterface, nil
	}
	cdi, err := c.client.GetResourceInterface(c.resGVK, c.resNamespace)
	if err != nil {
		return nil, err
	}
	c.resDynInterface = cdi
	return cdi, nil
}

func (c *TestA) refreshClient() (err error) {
	c.client, err = kgs.NewDynClient()
	return
}

func (c *TestA) refreshResDynInterface() (err error) {
	c.resDynInterface, err = c.client.GetResourceInterface(c.resGVK, c.resNamespace)
	return
}

func (c *TestA) refresh() error {
	fns := kgs.TestFns{
		c.refreshClient,
		c.refreshResDynInterface,
	}
	return fns.Run()
}

func (c *TestA) createCRD() error {
	ri, err := c.getDynamicInterfaceForCRD()
	if err != nil {
		return err
	}
	_, err = ri.Create(c.crd, metav1.CreateOptions{})
	return err
}

func (c *TestA) registerScheme() error {
	addKnownTypes := func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypeWithName(c.resGVK, &unstructured.Unstructured{})
		metav1.AddToGroupVersion(scheme, c.resGVK.GroupVersion())
		return nil
	}
	schemeBuilder := runtime.SchemeBuilder{addKnownTypes}

	schemeInst := runtime.NewScheme()
	serializer.NewCodecFactory(schemeInst)
	runtime.NewParameterCodec(schemeInst)
	metav1.AddToGroupVersion(schemeInst, schema.GroupVersion{Version: "v1"})
	return schemeBuilder.AddToScheme(schemeInst)
}

func (c *TestA) createA() error {
	ri, err := c.getDynamicInterfaceForRes()
	if err != nil {
		return err
	}
	_, err = ri.Create(c.resourceA, metav1.CreateOptions{})
	return err
}

func (c *TestA) createB() error {
	ri, err := c.getDynamicInterfaceForRes()
	if err != nil {
		return err
	}
	_, err = ri.Create(c.resourceB, metav1.CreateOptions{})
	return err
}

func (c *TestA) getAndMatchRes(given *unstructured.Unstructured) error {
	ri, err := c.getDynamicInterfaceForRes()
	if err != nil {
		return err
	}
	got, err := ri.Get(given.GetName(), metav1.GetOptions{})
	if err != nil {
		return err
	}
	if reflect.DeepEqual(given, got) {
		return nil
	}
	return errors.Errorf("failed match:\nexpected: %+v\ngot: %+v", given, got)
}

func (c *TestA) getAndMatchA() error {
	return c.getAndMatchRes(c.resourceA)
}

func (c *TestA) getAndMatchB() error {
	return c.getAndMatchRes(c.resourceB)
}

func (c *TestA) deleteCRD() error {
	ri, err := c.getDynamicInterfaceForCRD()
	if err != nil {
		return err
	}
	deletePropagation := metav1.DeletePropagationForeground
	return ri.Delete(
		c.crd.GetName(),
		&metav1.DeleteOptions{PropagationPolicy: &deletePropagation},
	)
}

func (c *TestA) verifyNoResInstances() error {
	ri, err := c.getDynamicInterfaceForRes()
	if err != nil {
		return err
	}
	_, err = ri.List(metav1.ListOptions{})
	if k8serrors.IsNotFound(err) {
		return nil
	}
	return err
}
