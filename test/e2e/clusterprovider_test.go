//go:generate opencontrolplane-gen
package e2e

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"text/template"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	// opencontrolplane-gen:replace github.com/openmcp-project/cluster-provider-template=MODULE
	"github.com/openmcp-project/cluster-provider-template/api/v1alpha1"
	clustersv1alpha1 "github.com/openmcp-project/openmcp-operator/api/clusters/v1alpha1"
	"github.com/openmcp-project/openmcp-operator/api/common"
	openmcpconditions "github.com/openmcp-project/openmcp-testing/pkg/conditions"
	"github.com/openmcp-project/openmcp-testing/pkg/providers"
)

const openmcpSystem = "openmcp-system"

func TestClusterProvider(t *testing.T) {
	basicClusterProviderTest := features.New("provider test").
		WithSetup("create provider config", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			v1alpha1.AddToScheme(c.Client().Resources().GetScheme())
			clustersv1alpha1.AddToScheme(c.Client().Resources().GetScheme())
			config := &v1alpha1.ProviderConfig{}
			// opencontrolplane-gen:replace configname=SERVICE_NAME
			config.SetName("configname")
			if err := c.Client().Resources().Create(ctx, config); err != nil {
				t.Errorf("failed to create ProviderConfig: %v", err)
			}
			return ctx
		}).
		Assess("verify cluster profiles have been created",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				clusterProfile := clustersv1alpha1.ClusterProfile{}
				// TODO replace with a cluster profile created by your cluster provided
				clusterProfile.Name = "kind"
				list := &clustersv1alpha1.ClusterProfileList{
					Items: []clustersv1alpha1.ClusterProfile{
						clusterProfile,
					},
				}
				if err := wait.For(conditions.New(c.Client().Resources()).ResourcesFound(list)); err != nil {
					t.Errorf("cluster profile not found: %v", err)
				}
				return ctx
			}).
		Assess("update cluster scheduler purpose mapping", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// TODO: replace with a profile that maps to your cluster provider
			addProfileToOpenMCPOperatorConfig(ctx, c.Client(), "openmcp-operator", openmcpSystem, "kind")
			return ctx
		}).
		Assess("verify control plane cluster request result in working cluster",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				clusterRequest := &clustersv1alpha1.ClusterRequest{}
				clusterRequest.SetName("test-cluster")
				clusterRequest.SetNamespace(openmcpSystem)
				clusterRequest.Spec.Purpose = "test"
				if err := c.Client().Resources().Create(ctx, clusterRequest); err != nil {
					t.Errorf("failed to create cluster request: %v", err)
					return ctx
				}
				cluster := &clustersv1alpha1.Cluster{}
				cluster.SetName("test")
				cluster.SetNamespace(openmcpSystem)
				if err := wait.For(openmcpconditions.Match(cluster, c, "Ready", corev1.ConditionTrue)); err != nil {
					t.Errorf("cluster is not ready")
				}
				return ctx
			}).
		Assess("verify access request result in kubeconfig for created control plane",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				accessRequest := &clustersv1alpha1.AccessRequest{}
				accessRequest.SetName("test")
				accessRequest.SetNamespace(openmcpSystem)
				accessRequest.Spec.ClusterRef = &common.ObjectReference{
					Name:      "test",
					Namespace: openmcpSystem,
				}
				accessRequest.Spec.Token = &clustersv1alpha1.TokenConfig{
					RoleRefs: []common.RoleRef{
						{
							Name: "cluster-admin",
							Kind: "ClusterRole",
						},
					},
				}
				if err := c.Client().Resources().Create(ctx, accessRequest); err != nil {
					t.Errorf("failed to created access request: %v", err)
					return ctx
				}
				if err := wait.For(openmcpconditions.Match(accessRequest, c, "Ready", corev1.ConditionTrue)); err != nil {
					t.Errorf("access request is not ready")
				}
				return ctx
			}).
		Assess("verify cluster is successfully deleted",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				if err := providers.DeleteCluster(ctx, c, types.NamespacedName{Namespace: openmcpSystem, Name: "test"}); err != nil {
					t.Errorf("delete cluster failed: %v", err)
				}
				return ctx
			})
	testenv.Test(t, basicClusterProviderTest.Feature())
}

func addProfileToOpenMCPOperatorConfig(ctx context.Context, c klient.Client, name, namespace, profileName string) error {
	cm := &corev1.ConfigMap{}
	if err := c.Resources().Get(ctx, name, namespace, cm); err != nil {
		return fmt.Errorf("failed to fetch ConfigMap %s/%s: %w", namespace, name, err)
	}
	data, ok := cm.Data["config"]
	if !ok {
		return fmt.Errorf("config key does not exist")
	}
	data, err := addPurposeMapping(purposeMapping{
		Purpose: "test",
		Profile: profileName,
	})
	if err != nil {
		return err
	}
	cm.Data["config"] = data
	if err := c.Resources().Update(ctx, cm); err != nil {
		return fmt.Errorf("failed to update ConfigMap %s/%s: %w", namespace, name, err)
	}
	// restart openmcp-operator to reload the config
	pods := &corev1.PodList{}
	if err := c.Resources().List(ctx, pods, resources.WithLabelSelector("app=openmcp-operator")); err != nil {
		return fmt.Errorf("failed to list openmcp-operator pod by label selector: %w", err)
	}
	for _, p := range pods.Items {
		c.Resources().Delete(ctx, &p)
	}
	return nil
}

type purposeMapping struct {
	Purpose string
	Profile string
}

func addPurposeMapping(mapping purposeMapping) (string, error) {
	tmpl, err := template.New("configTemplate").Parse(openmcpOperatorConfig)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}
	result := strings.Builder{}
	if err := tmpl.Execute(&result, mapping); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return result.String(), nil
}

const openmcpOperatorConfig = `
managedControlPlane:
  mcpClusterPurpose: mcp
scheduler:
  scope: Cluster
  purposeMappings:
    mcp:
      template:
        spec:
          profile: kind
          tenancy: Exclusive
    platform:
      template:
        spec:
          profile: kind
          tenancy: Shared
    onboarding:
      template:
        spec:
          profile: kind
          tenancy: Shared
    workload:
      template:
        spec:
          profile: kind
          tenancy: Shared
    {{.Purpose}}:
      template:
        spec:
          profile: {{.Profile}}
          tenancy: Shared
`
