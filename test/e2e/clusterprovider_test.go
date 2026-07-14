//go:generate opencontrolplane-gen
package e2e

import (
	"context"
	"fmt"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	// opencontrolplane-gen:replace github.com/openmcp-project/cluster-provider-template=MODULE
	"github.com/openmcp-project/cluster-provider-template/api/v1alpha1"
)

func TestClusterProvider(t *testing.T) {
	basicClusterProviderTest := features.New("provider test").
		WithSetup("create provider config", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			v1alpha1.AddToScheme(c.Client().Resources().GetScheme())
			config := &v1alpha1.ProviderConfig{}
			// opencontrolplane-gen:replace configname=SERVICE_NAME
			config.SetName("configname")
			if err := c.Client().Resources().Create(ctx, config); err != nil {
				t.Errorf("failed to create ProviderConfig object: %v", err)
			}
			return ctx
		}).
		Assess("verify cluster profiles have been created",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				// TODO
				return ctx
			}).
		Assess("update purpose mapping", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// replace test with actual profile name
			updateOpenMCPOperatorConfig(ctx, c.Client(), "openmcp-operator", "openmcp-system", "test")
			return ctx
		}).
		Assess("verify control plane cluster request result in working cluster",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				// TODO
				return ctx
			}).
		Assess("verify access request result in kubeconfig for created control plane",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				// TODO
				return ctx
			}).
		Assess("verify cluster is successfully deleted",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				// TODO
				return ctx
			})
	testenv.Test(t, basicClusterProviderTest.Feature())
}

func updateOpenMCPOperatorConfig(ctx context.Context, c klient.Client, name, namespace, profileName string) error {
	cm := &corev1.ConfigMap{}
	if err := c.Resources().Get(ctx, name, namespace, cm); err != nil {
		return fmt.Errorf("failed to fetch ConfigMap %s/%s: %w", namespace, name, err)
	}
	if err := updateProfile("test", cm); err != nil {
		return fmt.Errorf("failed to update profile in ConfigMap %s/%s: %w", namespace, name, err)
	}
	if err := c.Resources().Update(ctx, cm); err != nil {
		return fmt.Errorf("failed to update ConfigMap %s/%s: %w", namespace, name, err)
	}
	// restart openmcp-operator to reload the config
	pods := &corev1.PodList{}
	if err := c.Resources().List(ctx, pods, resources.WithLabelSelector("app=openmcp-operator")); err != nil {
		return fmt.Errorf("failed to list openmcp-operator pod by label selector: %w", err)
	}
	for _, p := range pods.Items {
		c.Resources().Delete(ctx, p)
	}
	return nil
}

func updateProfile(profileName string, cm *corev1.ConfigMap) error {
	data, ok := cm.Data["config"]
	if !ok {
		return fmt.Errorf("config key does not exist")
	}
	cm.Data["config"] = strings.Replace(data, "profile: kind", fmt.Sprintf("profile: %s", profileName), 1)
	fmt.Println(cm.Data["config"])
	return nil
}
