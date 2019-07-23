package kgetset

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

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
