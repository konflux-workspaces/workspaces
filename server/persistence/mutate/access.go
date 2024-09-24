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

import (
	"context"
	"strconv"

	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/persistence/clientinterface"
)

// Applies the has-direct-access label if the user "accessor" has access via
// public viewer or via a direct binding to the workspace
func ApplyHasDirectAccessLabel(
	ctx context.Context,
	client clientinterface.DirectAccessChecker,
	workspace *restworkspacesv1alpha1.Workspace,
	accessor string,
) error {
	if workspace == nil {
		return nil
	}
	if workspace.Labels == nil {
		workspace.Labels = map[string]string{}
	}

	ok, err := client.UserHasDirectAccess(ctx, accessor, workspace.Namespace)
	if err != nil {
		return err
	}

	workspace.Labels[restworkspacesv1alpha1.LabelHasDirectAccess] = strconv.FormatBool(ok)

	return nil
}
