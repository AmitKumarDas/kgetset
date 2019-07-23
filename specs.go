package kgetset

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func HelloCRD() *unstructured.Unstructured {
	return &unstructured.Unstructured{
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
}

func HelloResourceUn() *unstructured.Unstructured {
	return &unstructured.Unstructured{
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
