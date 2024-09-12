/*
Copyright 2024 The Workspaces Authors.

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

package mutate

import restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"

// Applies the is-owner internal label to a workspace
func ApplyIsOwnerLabel(workspace *restworkspacesv1alpha1.Workspace, owner string) {
	// do nothing on an empty workspace
	if workspace == nil {
		return
	}

	if workspace.Labels == nil {
		workspace.Labels = map[string]string{}
	}

	switch workspace.Namespace {
	case owner:
		workspace.Labels[restworkspacesv1alpha1.LabelIsOwner] = "true"
	default:
		workspace.Labels[restworkspacesv1alpha1.LabelIsOwner] = "false"
	}
}
