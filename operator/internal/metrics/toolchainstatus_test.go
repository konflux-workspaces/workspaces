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

package metrics_test

import (
	"context"
	_ "embed"
	"encoding/json"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/operator/internal/metrics"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:embed toolchainstatus_test.json
var goodToolchainStatus []byte

var _ = Describe("availabiltiy metrics", func() {
	name := "toolchain-status"
	namespace := "toolchain-host-operator"
	var clientBuilder *fake.ClientBuilder
	var scheme *runtime.Scheme
	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(toolchainv1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())

		clientBuilder = fake.NewClientBuilder().WithScheme(scheme)
	})

	When("toolchain status is not found", func() {
		It("should return an unready status", func() {
			gauge := metrics.NewToolchainStatusGauge(clientBuilder.Build(), namespace)
			Expect(gauge.Status(context.Background())).To(BeFalse())
		})
	})

	When("toolchain status is available", func() {
		var toolchainStatus toolchainv1alpha1.ToolchainStatus

		JustBeforeEach(func() {
			clientBuilder.WithObjects(&toolchainStatus)
		})

		When("toolchain status is empty", func() {
			BeforeEach(func() {
				toolchainStatus = toolchainv1alpha1.ToolchainStatus{}
			})
			It("should return an unready status", func() {
				gauge := metrics.NewToolchainStatusGauge(clientBuilder.Build(), namespace)
				Expect(gauge.Status(context.Background())).To(BeFalse())
			})
		})

		When("toolchain status is unready", func() {
			BeforeEach(func() {
				now := metav1.Now()
				toolchainStatus = toolchainv1alpha1.ToolchainStatus{
					ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
					Status: toolchainv1alpha1.ToolchainStatusStatus{
						HostOperator: &toolchainv1alpha1.HostOperatorStatus{
							Conditions: []toolchainv1alpha1.Condition{
								{
									Type:               toolchainv1alpha1.ConditionReady,
									Status:             v1.ConditionFalse,
									Reason:             "failure",
									LastTransitionTime: now,
									LastUpdatedTime:    &now,
								},
							},
						},
					},
				}
			})

			It("should return an unready status", func() {
				gauge := metrics.NewToolchainStatusGauge(clientBuilder.Build(), namespace)
				Expect(gauge.Status(context.Background())).To(BeFalse())
			})
		})

		When("toolchain status is ready", func() {
			BeforeEach(func() {
				Expect(json.Unmarshal(goodToolchainStatus, &toolchainStatus)).NotTo(HaveOccurred())
			})

			It("should return a ready status", func() {
				gauge := metrics.NewToolchainStatusGauge(clientBuilder.Build(), namespace)
				Expect(gauge.Status(context.Background())).To(BeTrue())
			})
		})
	})
})
