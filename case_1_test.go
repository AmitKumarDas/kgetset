package kgetset

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type key struct {
	name      string
	namespace string
}

type unclient struct {
	dynamic dynamic.Interface

	// Mapper is used to map GroupVersionKinds to Resources
	mapper meta.RESTMapper
}

func newUnClientOrDie() *unclient {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		panic(err)
	}
	gr, err := restmapper.GetAPIGroupResources(dc)
	if err != nil {
		panic(err)
	}
	return &unclient{
		dynamic: dyn,
		mapper:  restmapper.NewDiscoveryRESTMapper(gr),
	}
}

func (uc *unclient) getResourceInterface(gvk schema.GroupVersionKind, ns string) (dynamic.ResourceInterface, error) {
	mapping, err := uc.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	if mapping.Scope.Name() == meta.RESTScopeNameRoot {
		return uc.dynamic.Resource(mapping.Resource), nil
	}
	return uc.dynamic.Resource(mapping.Resource).Namespace(ns), nil
}

type testopts struct {
	key    key
	given  runtime.Object
	got    runtime.Object
	client *unclient
}

type k8sapi struct {
	call func(*testopts) error
}

func newK8sAPI(fn func(*testopts) error) k8sapi {
	return k8sapi{call: fn}
}

func create() k8sapi {
	return newK8sAPI(func(o *testopts) error {
		u, ok := o.given.(*unstructured.Unstructured)
		if !ok {
			return errors.Errorf(
				"create failed: given instance is not unstructured type: %T",
				o.given,
			)
		}

		ri, err := o.client.getResourceInterface(
			u.GroupVersionKind(),
			o.key.namespace,
		)
		if err != nil {
			return err
		}

		o.got, err = ri.Create(u, metav1.CreateOptions{})
		return err
	})
}

func get() k8sapi {
	return newK8sAPI(func(o *testopts) (err error) {
		u, ok := o.given.(*unstructured.Unstructured)
		if !ok {
			return errors.Errorf(
				"get failed: given instance is not unstructured type: %T",
				o.given,
			)
		}

		ri, err := o.client.getResourceInterface(
			u.GroupVersionKind(),
			o.key.namespace,
		)
		if err != nil {
			return err
		}

		o.got, err = ri.Get(o.key.name, metav1.GetOptions{})
		return
	})
}

type verifyopts struct {
	expect runtime.Object
	actual runtime.Object
	err    error
}

type verify struct {
	call func(o *verifyopts) bool
}

func newVerify(fn func(o *verifyopts) bool) verify {
	return verify{call: fn}
}

func noopVerify() verify {
	return newVerify(func(o *verifyopts) bool {
		return true
	})
}

func deepEqual() verify {
	return newVerify(func(o *verifyopts) bool {
		if reflect.DeepEqual(o.expect, o.actual) {
			return true
		}

		o.err = errors.Errorf(
			"unequal objects:\n%s",
			diff.ObjectGoPrintSideBySide(o.expect, o.actual),
		)
		return false
	})
}

type test struct {
	to *testopts
	vo *verifyopts
	k8sapi
	verify
}

func TestUseCaseA(tt *testing.T) {
	client := newUnClientOrDie()
	crdObj := HelloCRD()

	var steps = []string{
		"create_crd",
		"get_crd",
	}
	var tests = map[string]test{
		"create_crd": test{
			to: &testopts{
				key:    key{name: crdObj.GetName()},
				given:  crdObj,
				client: client,
			},
			k8sapi: create(),
			verify: noopVerify(),
		},

		"get_crd": test{
			to: &testopts{
				key: key{name: crdObj.GetName()},
				given: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind":       crdObj.GetKind(),
						"apiVersion": crdObj.GetAPIVersion(),
					},
				},
				client: client,
			},
			vo: &verifyopts{
				expect: crdObj,
			},
			k8sapi: get(),
			verify: deepEqual(),
		},
	}

	// test case needs to follow this order
	for _, step := range steps {
		t := tests[step]
		err := t.k8sapi.call(t.to)
		if err != nil {
			tt.Fatalf("test %q failed during preparation: %+v", step, err)
		}
		t.vo.actual = t.to.given
		if !t.verify.call(t.vo) {
			tt.Fatalf("test %q failed during verification: %+v", step, t.vo.err)
		}
	}
}
