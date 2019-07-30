/*
Copyright 2019 The MayaData Authors.
Copyright 2018 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package apply is a dynamic, client-side substitute for `kubectl apply` that
// tries to guess the right thing to do without any type-specific knowledge.
// Instead of generating a PATCH request, it does the patching locally and
// returns a full object with the ResourceVersion intact.
//
// We can't use actual `kubectl apply` yet because it doesn't support strategic
// merge for CRDs, which would make it infeasible to include a PodTemplateSpec
// in a CRD (e.g. containers and volumes will merge incorrectly).

// This entire piece of code is lifted from MetaController
//
// refer - https://github.com/GoogleCloudPlatform/metacontroller
package kgetset

import (
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
)

const (
	lastAppliedAnnotation = "dc.openebs.io/last-applied-state"
)

// SetLastApplied sets the last applied state against the provided object
func SetLastApplied(obj *unstructured.Unstructured, lastApplied map[string]interface{}) error {
	lastAppliedJSON, err := json.Marshal(lastApplied)
	if err != nil {
		return errors.Wrapf(err, "failed to set last applied state")
	}

	ann := obj.GetAnnotations()
	if ann == nil {
		ann = make(map[string]string, 1)
	}
	ann[lastAppliedAnnotation] = string(lastAppliedJSON)
	obj.SetAnnotations(ann)
	return nil
}

// GetLastApplied returns the last applied state from the provided object
func GetLastApplied(obj *unstructured.Unstructured) (map[string]interface{}, error) {
	lastAppliedJSON := obj.GetAnnotations()[lastAppliedAnnotation]
	if lastAppliedJSON == "" {
		return nil, nil
	}
	lastApplied := make(map[string]interface{})
	err := json.Unmarshal([]byte(lastAppliedJSON), &lastApplied)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to get last applied state from annotation %q",
			lastAppliedAnnotation,
		)
	}
	return lastApplied, nil
}

// Merge updates the given observed object to apply the desired changes.
// It returns an updated copy of the observed object if no error occurs.
//
//               last
//             applied
//                |
//               \|/
// desired --->(state)<--- observed
func Merge(observed, lastApplied, desired map[string]interface{}) (map[string]interface{}, error) {
	// Make a copy of observed since merge() mutates the observed.
	destination := runtime.DeepCopyJSON(observed)

	// we start the merge as a recursion by starting off with
	// an empty fieldPath
	if _, err := merge("", destination, lastApplied, desired); err != nil {
		return nil, errors.Wrapf(err, "failed to merge desired changes")
	}
	return destination, nil
}

// merge does the following:
//  1/ finds the diff from lastApplied to desired,
//  2/ then applies this diff against the observed,
//  3/ finally returning the replacement value for the destination.
//
// NOTE:
//  This logic is invoked recursively
func merge(
	fieldPath string,
	destination, lastApplied, desired interface{},
) (interface{}, error) {
	switch destVal := destination.(type) {
	case map[string]interface{}:
		// destination is an object
		// Make sure the lastApplied & desired are map[string]interface{} too
		lastVal, ok := lastApplied.(map[string]interface{})
		if !ok && lastVal != nil {
			return nil, errors.Errorf(
				"failed to merge: invalid lastApplied value for path %q: expecting map[string]interface{}, got %T",
				fieldPath,
				lastApplied,
			)
		}
		desVal, ok := desired.(map[string]interface{})
		if !ok && desVal != nil {
			return nil, errors.Errorf(
				"failed to merge: invalid desired value for path %q: expecting map[string]interface{}, got %T",
				fieldPath,
				desired,
			)
		}

		return mergeObject(fieldPath, destVal, lastVal, desVal)
	case []interface{}:
		// destination is an array
		// Make sure the lastApplied & desired are arrays too (or null).
		lastVal, ok := lastApplied.([]interface{})
		if !ok && lastVal != nil {
			return nil, errors.Errorf(
				"failed to merge: invalid lastApplied value for path %q: expecting []interface{}, got %T",
				fieldPath,
				lastApplied,
			)
		}
		desVal, ok := desired.([]interface{})
		if !ok && desVal != nil {
			return nil, errors.Errorf(
				"failed to merge: invalid desired value for path %q: expecting []interface, got %T",
				fieldPath,
				desired,
			)
		}

		return mergeArray(fieldPath, destVal, lastVal, desVal)
	default:
		// destination is a scalar or null; so desired value
		// is the final merge value
		return desired, nil
	}
}

// mergeObject deals specifically to merge objects i.e.
// map[string]interface{}
//
// NOTE:
//  Logic does a remove and then either an add or update
// of the keys to ensure a merge
func mergeObject(
	fieldPath string,
	destination, lastApplied, desired map[string]interface{},
) (interface{}, error) {
	// Remove fields that were present in lastApplied,
	// but are no longer desired
	for key := range lastApplied {
		if _, present := desired[key]; !present {
			delete(destination, key)
		}
	}

	// Add/Update all fields present in desired
	var err error
	for key, desVal := range desired {
		// need to invoke merge again since
		// we are not sure of the value's datatype
		destination[key], err = merge(
			fmt.Sprintf("%s[%s]", fieldPath, key),
			destination[key],
			lastApplied[key],
			desVal,
		)
		if err != nil {
			return nil, err
		}
	}

	return destination, nil
}

// mergeArray deals specifically to merge arrays i.e.
// []interface{}
//
// NOTE:
//  This is a special case since it is not possible to
// merge arrays without having extra information
//
//  Logic tries to detect if this list/array is a typical
// K8s based listOfMaps. If true, it tries to do a merge
// else returns the entire array of desired state to be used
// at destination. In other words, this logic does an array
// replacement for arrays that are not K8s based listOfMaps.
func mergeArray(
	fieldPath string,
	destination, lastApplied, desired []interface{},
) (interface{}, error) {
	// If it looks like a list map, use the special merge i.e. k8s based
	mergeKey := detectListMapKey(destination, lastApplied, desired)
	if mergeKey != "" {
		return mergeListMap(fieldPath, mergeKey, destination, lastApplied, desired)
	}

	// It's a normal array. Just make use of desired for now.
	// TODO(enisoc): Check if there are any common cases where
	// we want to merge.
	return desired, nil
}

// mergeListMap is a k8s specific case of mergeArray where each state
// is provided as an array of maps
func mergeListMap(
	fieldPath, mergeKey string,
	destination, lastApplied, desired []interface{},
) (interface{}, error) {
	// Treat each list of objects as if it were a map,
	// keyed by the mergeKey field.
	destMap := makeListMap(mergeKey, destination)
	lastMap := makeListMap(mergeKey, lastApplied)
	desMap := makeListMap(mergeKey, desired)

	// use merge logic that is applicable for map
	_, err := mergeObject(fieldPath, destMap, lastMap, desMap)
	if err != nil {
		return nil, err
	}

	// Turn destMap back into a list by **trying to preserve partial order**
	destList := make([]interface{}, 0, len(destMap))
	added := make(map[string]bool, len(destMap))

	// First consider merged items based on the destination list.
	for _, item := range destination {
		key := stringMergeKey(item.(map[string]interface{})[mergeKey])
		if newItem, present := destMap[key]; present {
			destList = append(destList, newItem)
			// Remember which items we've already added to the final list.
			added[key] = true
		}
	}

	// Then consider merged items based on the desired list
	for _, item := range desired {
		key := stringMergeKey(item.(map[string]interface{})[mergeKey])
		if !added[key] {
			destList = append(destList, destMap[key])
			added[key] = true
		}
	}

	return destList, nil
}

// makeListMap builds a map from the given list
func makeListMap(mergeKey string, list []interface{}) map[string]interface{} {
	res := make(map[string]interface{}, len(list))
	for _, item := range list {
		// NOTE:
		//  We only end up here if detectListMapKey() already verified
		// that all items are objects.
		itemMap := item.(map[string]interface{})
		res[stringMergeKey(itemMap[mergeKey])] = item
	}
	return res
}

// stringMergeKey converts a non string based merge key to its string
// equivalent.
func stringMergeKey(key interface{}) string {
	switch tkey := key.(type) {
	case string:
		return tkey
	default:
		return fmt.Sprintf("%v", key)
	}
}

// knownMergeKeys lists the key names we will guess as merge keys.
//
// The order determines precedence if multiple entries might work,
// with the first item having the highest precedence.
//
// Note that we don't do merges on status because the controller is solely
// responsible for providing the entire contents of status.
// As a result, we don't try to handle things like status.conditions.
var knownMergeKeys = []string{
	"containerPort",
	"port",
	"name",
	"uid",
	"ip",
}

// detectListMapKey tries to guess whether a field is a k8s-style "list map".
// e.g. k8s array of containers where each container is a map
//
// One provides all known examples of values for the field.
// If a **likely merge key** can be found, we return it.
// Otherwise, we return an empty string.
//
// NOTE:
//  Following is the overall idea to enable merging a list:
//  1/ filter the common keys amongst all the items
//  2/ find the most likely merge key based on a known list of keys
func detectListMapKey(lists ...[]interface{}) string {
	// Remember the set of keys that every object has in common.
	var commonKeys map[string]bool

	for _, list := range lists {
		for _, item := range list {
			// All the items must be objects i.e. k8s maps
			obj, ok := item.(map[string]interface{})
			if !ok {
				return ""
			}

			// Initialize commonKeys to the keys of the first object seen
			// NOTE: This block is executed only once
			if commonKeys == nil {
				commonKeys = make(map[string]bool, len(obj))
				for key := range obj {
					commonKeys[key] = true
				}
				// continue to next item in this initialize block
				// else it gets pruned
				continue
			}

			// For all other objects, prune the set.
			for key := range commonKeys {
				if _, present := obj[key]; !present {
					delete(commonKeys, key)
				}
			}
		}
	}

	// If all objects have one of the known conventional merge keys in common,
	// we'll guess that this is a list map
	for _, key := range knownMergeKeys {
		if commonKeys[key] {
			return key
		}
	}
	return ""
}
