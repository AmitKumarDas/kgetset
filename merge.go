/*
Copyright 2019 The MayaData Authors
Copyright 2019 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kgetset

import (
	"encoding/json"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// IsDiff finds if original & current objects differ
// by checking if a patch file gets formed out of the
// provided instances
//
// NOTE:
//  Original is the original object (last-applied config in annotation)
//  Modified is the modified object (new config we want)
//  Current is the current object (live config in the server)
//
// NOTE:
//  For the purpose of calculating diff Modified is same as Current
func IsDiff(original, current *unstructured.Unstructured) (bool, error) {
	if original == nil || current == nil {
		return false, errors.Errorf(
			"failed to diff: instances can not be nil:\noriginal: '%+v'\ncurrent: '%+v'",
			original,
			current,
		)
	}

	// We need JSON bytes to generate a patch so marshal the original
	origAsJSON, err := json.Marshal(original)
	if err != nil {
		return false, err
	}

	// We need JSON bytes to generate a patch so marshal the current
	currentAsJSON, err := json.Marshal(current)
	if err != nil {
		return false, err
	}

	// Get the patch meta for unstruct, which is needed for generating
	// and applying the merge patch.
	patchSchema, err := strategicpatch.NewPatchMetaFromStruct(current)
	if err != nil {
		return false, err
	}

	// Create a merge patch
	patch, err := strategicpatch.CreateTwoWayMergePatch(
		origAsJSON,
		currentAsJSON,
		patchSchema,
	)
	if err != nil {
		return false, err
	}

	// understand here
	return false, errors.Errorf("patch:\n%+v", patch)
}
