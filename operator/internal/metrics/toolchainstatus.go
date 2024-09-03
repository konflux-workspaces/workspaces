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

package metrics

import (
	"context"
	"slices"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

//+kubebuilder:rbac:groups=toolchain.dev.openshift.com,resources=toolchainstatuses,verbs=get;list;watch

// ToolchainStatusGauge measures whether the underlying kubesaw instance is
// alive and running correctly.
type ToolchainStatusGauge struct {
	client.Client
	kubesawNamespace string
}

func NewToolchainStatusGauge(client client.Client, kubesawNamespace string) ToolchainStatusGauge {
	return ToolchainStatusGauge{Client: client, kubesawNamespace: kubesawNamespace}
}

const KonfluxWorkspacesAvailable = "konflux_workspaces_available"
const KubesawToolchainStatusName = "toolchain-status"

func (r *ToolchainStatusGauge) Register(ctx context.Context) {
	KubesawGauge := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: KonfluxWorkspacesAvailable,
	}, func() float64 {
		if r.Status(ctx) {
			return 1.0
		}
		return 0.0
	})

	metrics.Registry.MustRegister(KubesawGauge)
}

func (r *ToolchainStatusGauge) Status(ctx context.Context) bool {
	toolchainstatus := toolchainv1alpha1.ToolchainStatus{}

	err := r.Client.Get(ctx,
		types.NamespacedName{Namespace: r.kubesawNamespace, Name: KubesawToolchainStatusName},
		&toolchainstatus)
	if err != nil {
		return false
	}

	conditions := slices.Concat(
		toolchainstatus.Status.Conditions,
		toolchainstatus.Status.HostRoutes.Conditions,
	)

	if toolchainstatus.Status.HostOperator != nil {
		conditions = slices.Concat(conditions,
			toolchainstatus.Status.HostOperator.Conditions,
			toolchainstatus.Status.HostOperator.RevisionCheck.Conditions)
	}

	if toolchainstatus.Status.RegistrationService != nil {
		conditions = slices.Concat(conditions,
			toolchainstatus.Status.RegistrationService.Health.Conditions,
			toolchainstatus.Status.RegistrationService.Deployment.Conditions,
			toolchainstatus.Status.RegistrationService.RevisionCheck.Conditions)
	}

	// we ignore Che in member clusters since we don't care about Che in konflux
	for _, member := range toolchainstatus.Status.Members {
		conditions = append(conditions, member.MemberStatus.Conditions...)
		if member.MemberStatus.Host != nil {
			conditions = append(conditions, member.MemberStatus.Host.Conditions...)
		}
		if member.MemberStatus.HostConnection != nil {
			conditions = append(conditions, member.MemberStatus.HostConnection.Conditions...)
		}
		if member.MemberStatus.Routes != nil {
			conditions = append(conditions, member.MemberStatus.Routes.Conditions...)
		}
	}

	return conditions != nil &&
		!slices.ContainsFunc(conditions, func(cond toolchainv1alpha1.Condition) bool {
			return cond.Type == toolchainv1alpha1.ConditionReady &&
				cond.Status != v1.ConditionTrue
		})
}
