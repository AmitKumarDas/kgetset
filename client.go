package kgetset

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type DynClient struct {
	// addToSchemes holds a list of registations
	// that need to be done against the scheme
	//addToSchemes []func(scheme *runtime.Scheme)

	dynamic dynamic.Interface

	// Mapper is used to map GroupVersionKinds to Resources
	mapper meta.RESTMapper
}

func NewDynClient() (*DynClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	gr, err := restmapper.GetAPIGroupResources(dc)
	if err != nil {
		return nil, err
	}
	return &DynClient{
		dynamic: dyn,
		mapper:  restmapper.NewDiscoveryRESTMapper(gr),
	}, nil
}

func NewDynClientOrDie() *DynClient {
	d, err := NewDynClient()
	if err != nil {
		panic(err)
	}
	return d
}

func (uc *DynClient) GetResourceInterface(
	gvk schema.GroupVersionKind,
	ns ...string,
) (dynamic.ResourceInterface, error) {
	mapping, err := uc.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	if mapping.Scope.Name() == meta.RESTScopeNameRoot {
		return uc.dynamic.Resource(mapping.Resource), nil
	}
	if len(ns) == 0 {
		return nil, errors.Errorf(
			"failed to get dynamic interface: missing namespace",
		)
	}
	ns0 := ns[0]
	return uc.dynamic.Resource(mapping.Resource).Namespace(ns0), nil
}
