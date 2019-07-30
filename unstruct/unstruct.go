package unstruct

import (
  "strings"

  "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func IsChangeStr(src, dest *unstructured.Unstructured, path string, others ...string) (bool, error) {
  allpaths := append(others, path)
  if len(allpaths) == 0 {
    return false, errors.New("failed to determine ischange: missing paths")
  }
  for _, p := range allpaths {
    changed, err := isChangeStr(src, dest, p)
    if err != nil {
      return false, err
    }
    if changed {
      return true, nil
    }
  }
  return false, nil
}

func isChangeStr(src, dest *unstructured.Unstructured, path string) (bool, error) {
  sval, _, err := unstructured.NestedString(
    src.Object, 
    strings.Split(path, ".")...,
  )
  if err != nil {
    return false, err
  }
  dval, _, err := unstructured.NestedString(
    dest.Object, 
    strings.Split(path, ".")...,
  )
  if err != nil {
    return false, err
  }
  return sval != dval, nil
}
