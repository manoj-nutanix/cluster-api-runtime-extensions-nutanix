// Copyright 2024 Nutanix. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package nutanix

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	caaphv1 "github.com/nutanix-cloud-native/cluster-api-runtime-extensions-nutanix/api/external/sigs.k8s.io/cluster-api-addon-provider-helm/api/v1alpha1"
	"github.com/nutanix-cloud-native/cluster-api-runtime-extensions-nutanix/common/pkg/k8s/client"
	"github.com/nutanix-cloud-native/cluster-api-runtime-extensions-nutanix/pkg/handlers/generic/lifecycle/config"
	handlersutils "github.com/nutanix-cloud-native/cluster-api-runtime-extensions-nutanix/pkg/handlers/utils"
)

type helmAddonConfig struct {
	defaultValuesTemplateConfigMapName string
}

func (c *helmAddonConfig) AddFlags(prefix string, flags *pflag.FlagSet) {
	flags.StringVar(
		&c.defaultValuesTemplateConfigMapName,
		prefix+".default-values-template-configmap-name",
		"default-nutanix-csi-helm-values-template",
		"default values ConfigMap name",
	)
}

type helmAddonStrategy struct {
	config            helmAddonConfig
	client            ctrlclient.Client
	helmChart         *config.HelmChart
	snapshotHelmChart *config.HelmChart
}

func (s helmAddonStrategy) apply(
	ctx context.Context,
	cluster *clusterv1.Cluster,
	defaultsNamespace string,
	log logr.Logger,
) error {
	log.Info("Retrieving Nutanix CSI installation values template for cluster")
	storageChartValues, err := handlersutils.RetrieveValuesTemplate(
		ctx,
		s.client,
		s.config.defaultValuesTemplateConfigMapName,
		defaultsNamespace,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to retrieve nutanix csi installation values template ConfigMap for cluster: %w",
			err,
		)
	}

	storageChartProxy := &caaphv1.HelmChartProxy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: caaphv1.GroupVersion.String(),
			Kind:       "HelmChartProxy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cluster.Namespace,
			Name:      "nutanix-csi-" + cluster.Name,
		},
		Spec: caaphv1.HelmChartProxySpec{
			RepoURL:   s.helmChart.Repository,
			ChartName: s.helmChart.Name,
			ClusterSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{clusterv1.ClusterNameLabel: cluster.Name},
			},
			ReleaseNamespace: defaultStorageHelmReleaseNamespace,
			ReleaseName:      defaultStorageHelmReleaseName,
			Version:          s.helmChart.Version,
			ValuesTemplate:   storageChartValues,
		},
	}
	handlersutils.SetTLSConfigForHelmChartProxyIfNeeded(storageChartProxy)

	if err = controllerutil.SetOwnerReference(cluster, storageChartProxy, s.client.Scheme()); err != nil {
		return fmt.Errorf(
			"failed to set owner reference on HelmChartProxy %q: %w",
			storageChartProxy.Name,
			err,
		)
	}

	if err = client.ServerSideApply(ctx, s.client, storageChartProxy, client.ForceOwnership); err != nil {
		return fmt.Errorf("failed to apply HelmChartProxy %q: %w", storageChartProxy.Name, err)
	}

	snapshotChartProxy := &caaphv1.HelmChartProxy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: caaphv1.GroupVersion.String(),
			Kind:       "HelmChartProxy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cluster.Namespace,
			Name:      "nutanix-csi-snapshot-" + cluster.Name,
		},
		Spec: caaphv1.HelmChartProxySpec{
			RepoURL:   s.snapshotHelmChart.Repository,
			ChartName: s.snapshotHelmChart.Name,
			ClusterSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{clusterv1.ClusterNameLabel: cluster.Name},
			},
			ReleaseNamespace: defaultSnapshotHelmReleaseNamespace,
			ReleaseName:      defaultSnapshotHelmReleaseName,
			Version:          s.snapshotHelmChart.Version,
		},
	}
	handlersutils.SetTLSConfigForHelmChartProxyIfNeeded(snapshotChartProxy)

	if err = controllerutil.SetOwnerReference(cluster, snapshotChartProxy, s.client.Scheme()); err != nil {
		return fmt.Errorf(
			"failed to set owner reference on HelmChartProxy %q: %w",
			snapshotChartProxy.Name,
			err,
		)
	}

	if err = client.ServerSideApply(ctx, s.client, snapshotChartProxy, client.ForceOwnership); err != nil {
		return fmt.Errorf("failed to apply HelmChartProxy %q: %w", snapshotChartProxy.Name, err)
	}

	return nil
}