//go:generate opencontrolplane-gen
package e2e

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	openmcpconditions "github.com/openmcp-project/openmcp-testing/pkg/conditions"

	"github.com/openmcp-project/openmcp-testing/pkg/resources"

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
		Assess("verify service can be successfully consumed",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				v1alpha1.AddToScheme(c.Client().Resources().GetScheme())
				// opencontrolplane-gen:replace Foo=KIND
				api := &v1alpha1.Foo{}
				api.SetName("test")
				api.SetNamespace(metav1.NamespaceDefault)
				if err := c.Client().Resources().Create(ctx, api); err != nil {
					// opencontrolplane-gen:replace Foo=KIND
					t.Errorf("failed to create Foo object: %v", err)
				}
				if err := wait.For(openmcpconditions.Match(api, c, "Ready", corev1.ConditionTrue)); err != nil {
					t.Error(err)
				}
				return ctx
			}).
		Assess("verify service can be successfully deleted",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				v1alpha1.AddToScheme(c.Client().Resources().GetScheme())
				// opencontrolplane-gen:replace Foo=KIND
				apiList := &v1alpha1.FooList{}
				if err := c.Client().Resources().List(ctx, apiList); err != nil {
					t.Error(err)
					return ctx
				}
				for _, obj := range apiList.Items {
					if err := resources.DeleteObject(ctx, c, &obj, wait.WithTimeout(time.Minute)); err != nil {
						// opencontrolplane-gen:replace Foo=KIND
						t.Errorf("failed to delete Foo object: %v", err)
					}
				}
				return ctx
			})
	testenv.Test(t, basicClusterProviderTest.Feature())
}
