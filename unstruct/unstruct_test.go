package unstruct

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var testUnstructInstA *unstructured.Unstructured = &unstructured.Unstructured{
	Object: map[string]interface{}{
		"kind":       "CustomResourceDefinition",
		"apiVersion": "apiextensions.k8s.io/v1beta1",
		"metadata": map[string]interface{}{
			"name": "hellos.openebs.io",
		},
	},
}

var testUnstructInstB *unstructured.Unstructured = &unstructured.Unstructured{
	Object: map[string]interface{}{
		"kind":       "CustomResourceDefinition",
		"apiVersion": "apiextensions.k8s.io/v1beta1",
		"metadata": map[string]interface{}{
			"name": "zellos.openebs.io",
		},
	},
}

var testUnstructInstC *unstructured.Unstructured = &unstructured.Unstructured{
	Object: map[string]interface{}{
		"kind":       "CustomResourceDefinition",
		"apiVersion": "apiextensions.k8s.io/v1beta1",
		"metadata": map[string]interface{}{
			"name": "byes.openebs.io",
		},
		"spec": map[string]interface{}{
			"group":   "openebs.io",
			"version": "v1",
			"scope":   "Namespaced",
			"names": map[string]interface{}{
				"plural":     "byes",
				"singular":   "bye",
				"kind":       "Bye",
				"shortNames": []string{"bye"},
			},
		},
	},
}

var testUnstructInstD *unstructured.Unstructured = &unstructured.Unstructured{
	Object: map[string]interface{}{
		"kind":       "CustomResourceDefinition",
		"apiVersion": "apiextensions.k8s.io/v1beta1",
		"metadata": map[string]interface{}{
			"name": "his.openebs.io",
		},
		"spec": map[string]interface{}{
			"group":   "openebs.io",
			"version": "v1",
			"scope":   "Namespaced",
			"names": map[string]interface{}{
				"plural":     "his",
				"singular":   "hi",
				"kind":       "Hi",
				"shortNames": []string{"hi"},
			},
		},
	},
}

func TestIsChangeStrWithSameName(t *testing.T) {
	changed, err := IsChangeStr(testUnstructInstA, testUnstructInstA, "metadata.name")
	if err != nil {
		t.Fatalf("test failed: %+v", err)
	}
	if changed {
		t.Fatalf("test failed: expected no change got change")
	}
}

func TestIsChangeStrWithDiffName(t *testing.T) {
	changed, err := IsChangeStr(testUnstructInstA, testUnstructInstB, "metadata.name")
	if err != nil {
		t.Fatalf("test failed: %+v", err)
	}
	if !changed {
		t.Fatalf("test failed: expected change got no change")
	}
}

func TestIsChangeStrWithSameGroup(t *testing.T) {
	changed, err := IsChangeStr(testUnstructInstC, testUnstructInstD, "spec.group")
	if err != nil {
		t.Fatalf("test failed: %+v", err)
	}
	if changed {
		t.Fatalf("test failed: expected no change got change")
	}
}

func TestIsChangeStrWithSameNamespace(t *testing.T) {
	changed, err := IsChangeStr(testUnstructInstC, testUnstructInstD, "spec.namespace")
	if err != nil {
		t.Fatalf("test failed: %+v", err)
	}
	if changed {
		t.Fatalf("test failed: expected no change got change")
	}
}

func TestIsChangeStrWithSameSingularName(t *testing.T) {
	changed, err := IsChangeStr(testUnstructInstC, testUnstructInstC, "spec.names.singular")
	if err != nil {
		t.Fatalf("test failed: %+v", err)
	}
	if changed {
		t.Fatalf("test failed: expected no change got change")
	}
}

func TestIsChangeStrWithSameEverything(t *testing.T) {
	changed, err := IsChangeStr(
		testUnstructInstC,
		testUnstructInstC,
		"metadata.name",
		"spec.group",
		"spec.version",
		"spec.scope",
		"spec.names.plural",
		"spec.names.singular",
		"spec.names.kind",
	)
	if err != nil {
		t.Fatalf("test failed: %+v", err)
	}
	if changed {
		t.Fatalf("test failed: expected no change got change")
	}
}
