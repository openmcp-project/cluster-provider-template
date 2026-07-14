//go:generate opencontrolplane-gen
package e2e

import (
	"context"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	// opencontrolplane-gen:replace github.com/openmcp-project/cluster-provider-template=MODULE
	"github.com/openmcp-project/cluster-provider-template/api/v1alpha1"
)

func TestClusterProvider(t *testing.T) {
	basicClusterProviderTest := features.New("provider test").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			v1alpha1.AddToScheme(c.Client().Resources().GetScheme())
			config := &v1alpha1.ProviderConfig{}
			// opencontrolplane-gen:replace configname=SERVICE_NAME
			config.SetName("configname")
			if err := c.Client().Resources().Create(ctx, config); err != nil {
				t.Errorf("failed to create ProviderConfig object: %v", err)
			}
			return ctx
		}).
		Assess("verify cluster requests result in real cluster",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				// TODO
				return ctx
			}).
		Assess("verify access requests result in kubeconfig",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				// TODO
				return ctx
			}).
		Assess("verify cluster can be successfully deleted",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				// TODO
				return ctx
			})
	testenv.Test(t, basicClusterProviderTest.Feature())
}
