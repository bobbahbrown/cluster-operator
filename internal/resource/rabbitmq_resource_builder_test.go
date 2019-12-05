package resource_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	rabbitmqv1beta1 "github.com/pivotal/rabbitmq-for-kubernetes/api/v1beta1"
	"github.com/pivotal/rabbitmq-for-kubernetes/internal/resource"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	defaultscheme "k8s.io/client-go/kubernetes/scheme"
)

var _ = Describe("RabbitmqResourceBuilder", func() {
	Context("ResourceBuilders", func() {
		var (
			instance = rabbitmqv1beta1.RabbitmqCluster{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "namespace",
				},
			}

			rabbitmqCluster *resource.RabbitmqResourceBuilder
			scheme          *runtime.Scheme
		)

		BeforeEach(func() {
			scheme = runtime.NewScheme()
			Expect(rabbitmqv1beta1.AddToScheme(scheme)).To(Succeed())
			Expect(defaultscheme.AddToScheme(scheme)).To(Succeed())
			rabbitmqCluster = &resource.RabbitmqResourceBuilder{
				Instance: &instance,
				DefaultConfiguration: resource.DefaultConfiguration{
					PersistentStorageClassName: "standard",
					PersistentStorage:          "10Gi",
					Scheme:                     scheme},
			}
		})

		When("no operator registry secret is set in the default configuration", func() {
			BeforeEach(func() {
				rabbitmqCluster.DefaultConfiguration.OperatorRegistrySecret = nil
			})
			It("returns the required resource builders in the expected order", func() {
				resourceBuilders, err := rabbitmqCluster.ResourceBuilders()
				Expect(err).NotTo(HaveOccurred())

				Expect(len(resourceBuilders)).To(Equal(8))

				resourceMap := checkForResourceBuilders(resourceBuilders)

				expectedKeys := []string{
					"0 - Service:test-rabbitmq-headless",
					"1 - Service:test-rabbitmq-ingress",
					"2 - Secret:test-rabbitmq-erlang-cookie",
					"3 - Secret:test-rabbitmq-admin",
					"4 - ConfigMap:test-rabbitmq-server-conf",
					"5 - ServiceAccount:test-rabbitmq-server",
					"6 - Role:test-rabbitmq-endpoint-discovery",
					"7 - RoleBinding:test-rabbitmq-server",
				}

				for index := range expectedKeys {
					Expect(resourceMap[expectedKeys[index]]).Should(BeTrue())
				}
			})
		})

		When("an operator Registry secret is set in the default configuration", func() {
			BeforeEach(func() {
				rabbitmqCluster.DefaultConfiguration.OperatorRegistrySecret = &corev1.Secret{}
			})

			It("returns the registry secret resource builder in the expected order", func() {
				resourceBuilders, err := rabbitmqCluster.ResourceBuilders()
				Expect(err).NotTo(HaveOccurred())

				Expect(len(resourceBuilders)).To(Equal(9))

				resourceMap := checkForResourceBuilders(resourceBuilders)

				expectedKeys := []string{
					"0 - Service:test-rabbitmq-headless",
					"1 - Service:test-rabbitmq-ingress",
					"2 - Secret:test-rabbitmq-erlang-cookie",
					"3 - Secret:test-rabbitmq-admin",
					"4 - ConfigMap:test-rabbitmq-server-conf",
					"5 - ServiceAccount:test-rabbitmq-server",
					"6 - Role:test-rabbitmq-endpoint-discovery",
					"7 - RoleBinding:test-rabbitmq-server",
					"8 - Secret:test-registry-access",
				}

				for index := range expectedKeys {
					Expect(resourceMap[expectedKeys[index]]).Should(BeTrue())
				}
			})
		})
	})
})

func checkForResources(resources []runtime.Object) (resourceMap map[string]bool) {
	resourceMap = make(map[string]bool)
	for i, resource := range resources {
		switch r := resource.(type) {
		case *corev1.Secret:
			if r.Name == "test-rabbitmq-admin" {
				resourceMap[fmt.Sprintf("%d - Secret:%s", i, r.Name)] = true
			}
			if r.Name == "test-rabbitmq-erlang-cookie" {
				resourceMap[fmt.Sprintf("%d - Secret:%s", i, r.Name)] = true
			}
			if r.Name == "test-registry-access" {
				resourceMap[fmt.Sprintf("%d - Secret:%s", i, r.Name)] = true
			}
		case *corev1.Service:
			if r.Name == "test-rabbitmq-headless" {
				resourceMap[fmt.Sprintf("%d - Service:%s", i, r.Name)] = true
			}
			if r.Name == "test-rabbitmq-ingress" {
				resourceMap[fmt.Sprintf("%d - Service:%s", i, r.Name)] = true
			}
		case *corev1.ConfigMap:
			if r.Name == "test-rabbitmq-server-conf" {
				resourceMap[fmt.Sprintf("%d - ConfigMap:%s", i, r.Name)] = true
			}
		case *corev1.ServiceAccount:
			if r.Name == "test-rabbitmq-server" {
				resourceMap[fmt.Sprintf("%d - ServiceAccount:%s", i, r.Name)] = true
			}
		case *rbacv1.Role:
			if r.Name == "test-rabbitmq-endpoint-discovery" {
				resourceMap[fmt.Sprintf("%d - Role:%s", i, r.Name)] = true
			}
		case *rbacv1.RoleBinding:
			if r.Name == "test-rabbitmq-server" {
				resourceMap[fmt.Sprintf("%d - RoleBinding:%s", i, r.Name)] = true
			}
		case *appsv1.StatefulSet:
			if r.Name == "test-rabbitmq-server" {
				resourceMap[fmt.Sprintf("%d - StatefulSet:%s", i, r.Name)] = true
			}
		}
	}
	return resourceMap
}

func checkForResourceBuilders(builders []resource.ResourceBuilder) (resourceMap map[string]bool) {
	resourceMap = make(map[string]bool)
	for i, builder := range builders {
		resource, _ := builder.Build()
		switch r := resource.(type) {
		case *corev1.Secret:
			if r.Name == "test-rabbitmq-admin" {
				resourceMap[fmt.Sprintf("%d - Secret:%s", i, r.Name)] = true
			}
			if r.Name == "test-rabbitmq-erlang-cookie" {
				resourceMap[fmt.Sprintf("%d - Secret:%s", i, r.Name)] = true
			}
			if r.Name == "test-registry-access" {
				resourceMap[fmt.Sprintf("%d - Secret:%s", i, r.Name)] = true
			}
		case *corev1.Service:
			if r.Name == "test-rabbitmq-headless" {
				resourceMap[fmt.Sprintf("%d - Service:%s", i, r.Name)] = true
			}
			if r.Name == "test-rabbitmq-ingress" {
				resourceMap[fmt.Sprintf("%d - Service:%s", i, r.Name)] = true
			}
		case *corev1.ConfigMap:
			if r.Name == "test-rabbitmq-server-conf" {
				resourceMap[fmt.Sprintf("%d - ConfigMap:%s", i, r.Name)] = true
			}
		case *corev1.ServiceAccount:
			if r.Name == "test-rabbitmq-server" {
				resourceMap[fmt.Sprintf("%d - ServiceAccount:%s", i, r.Name)] = true
			}
		case *rbacv1.Role:
			if r.Name == "test-rabbitmq-endpoint-discovery" {
				resourceMap[fmt.Sprintf("%d - Role:%s", i, r.Name)] = true
			}
		case *rbacv1.RoleBinding:
			if r.Name == "test-rabbitmq-server" {
				resourceMap[fmt.Sprintf("%d - RoleBinding:%s", i, r.Name)] = true
			}
		case *appsv1.StatefulSet:
			if r.Name == "test-rabbitmq-server" {
				resourceMap[fmt.Sprintf("%d - StatefulSet:%s", i, r.Name)] = true
			}
		}
	}
	return resourceMap
}
