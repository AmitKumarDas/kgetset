package onegvkdiffschemas

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var crdInst *unstructured.Unstructured = &unstructured.Unstructured{
	Object: map[string]interface{}{
		"kind":       "CustomResourceDefinition",
		"apiVersion": "apiextensions.k8s.io/v1beta1",
		"metadata": map[string]interface{}{
			"name": "onlyones.openebs.io",
		},
		"spec": map[string]interface{}{
			"group":   "openebs.io",
			"version": "v1alpha1",
			"scope":   "Namespaced",
			"names": map[string]interface{}{
				"plural":     "onlyones",
				"singular":   "onlyone",
				"kind":       "Onlyone",
				"shortNames": []string{"onlyone"},
			},
		},
	},
}

var resourceInstA *unstructured.Unstructured = &unstructured.Unstructured{
	Object: map[string]interface{}{
		"kind":       "Onlyone",
		"apiVersion": "v1alpha1",
		"metadata": map[string]interface{}{
			"name":      "onlyone-a",
			"namespace": "default",
			"labels": map[string]string{
				"app": "testing",
			},
		},
		"spec": map[string]interface{}{
			"count": 1,
			"desc":  "this is one",
			"id":    123,
		},
		"status": map[string]interface{}{
			"phase": "Up",
		},
	},
}

var resourceInstB *unstructured.Unstructured = &unstructured.Unstructured{
	Object: map[string]interface{}{
		"kind":       "Onlyone",
		"apiVersion": "v1alpha1",
		"metadata": map[string]interface{}{
			"name":      "onlyone-b",
			"namespace": "default",
			"labels": map[string]string{
				"app": "testing",
			},
		},
		"spec": map[string]interface{}{
			"count": 2,
			"desc":  "this is two",
			"id":    123,
			"addon": "enjoy",
		},
		"status": map[string]interface{}{
			"phase": "Down",
		},
	},
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Onlyone struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OnlyoneSpec   `json:"spec,omitempty"`
	Status OnlyoneStatus `json:"status,omitempty"`
}

type OnlyoneSpec struct {
	Count int    `json:"count"`
	Desc  string `json:"desc"`
	ID    string `json:"id"`
	Addon string `json:"addon"`
}

type OnlyoneStatus struct {
	Phase string `json:"phase"`
}
