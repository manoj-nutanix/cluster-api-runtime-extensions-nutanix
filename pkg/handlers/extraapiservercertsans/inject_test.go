// Copyright 2023 D2iQ, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package extraapiservercertsans

import (
	"testing"

	. "github.com/onsi/gomega"
	runtimehooksv1 "sigs.k8s.io/cluster-api/exp/runtime/hooks/api/v1alpha1"

	"github.com/d2iq-labs/capi-runtime-extensions/api/v1alpha1"
	"github.com/d2iq-labs/capi-runtime-extensions/common/pkg/capi/clustertopology/handlers/mutation"
	"github.com/d2iq-labs/capi-runtime-extensions/common/pkg/testutils/capitest"
)

func TestGeneratePatches(t *testing.T) {
	capitest.ValidateGeneratePatches(
		t,
		func() mutation.GeneratePatches { return NewPatch(VariableName) },
		capitest.PatchTestDef{
			Name: "unset variable",
		},
		capitest.PatchTestDef{
			Name: "extra API server cert SANs set",
			Vars: []runtimehooksv1.Variable{
				capitest.VariableWithValue(
					VariableName,
					v1alpha1.ExtraAPIServerCertSANs{"a.b.c.example.com", "d.e.f.example.com"},
				),
			},
			RequestItem: capitest.NewKubeadmControlPlaneTemplateRequestItem(),
			ExpectedPatchMatchers: []capitest.JSONPatchMatcher{{
				Operation: "add",
				Path:      "/spec/template/spec/kubeadmConfigSpec/clusterConfiguration",
				ValueMatcher: HaveKeyWithValue(
					"apiServer",
					HaveKeyWithValue(
						"certSANs",
						[]interface{}{"a.b.c.example.com", "d.e.f.example.com"},
					),
				),
			}},
		},
	)
}