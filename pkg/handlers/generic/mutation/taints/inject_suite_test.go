// Copyright 2023 Nutanix. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package taints

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTaintsPatch(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Taints patches for ControlPlane and Workers suite")
}
