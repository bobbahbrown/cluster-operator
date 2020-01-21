package resource_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	rabbitmqv1beta1 "github.com/pivotal/rabbitmq-for-kubernetes/api/v1beta1"
	"github.com/pivotal/rabbitmq-for-kubernetes/internal/resource"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	defaultscheme "k8s.io/client-go/kubernetes/scheme"
)

var _ = Context("IngressServices", func() {
	var (
		instance   rabbitmqv1beta1.RabbitmqCluster
		rmqBuilder resource.RabbitmqResourceBuilder
		scheme     *runtime.Scheme
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(rabbitmqv1beta1.AddToScheme(scheme)).To(Succeed())
		Expect(defaultscheme.AddToScheme(scheme)).To(Succeed())
		instance = rabbitmqv1beta1.RabbitmqCluster{}
		instance.Namespace = "foo"
		instance.Name = "foo"
		rmqBuilder = resource.RabbitmqResourceBuilder{
			Instance: &instance,
			Scheme:   scheme,
		}
	})

	It("generates Ingress Service with defaults", func() {
		serviceBuilder := rmqBuilder.IngressService()
		obj, err := serviceBuilder.Build()
		Expect(err).NotTo(HaveOccurred())
		service := obj.(*corev1.Service)

		By("generates a service object with the correct name and labels", func() {
			expectedName := instance.ChildResourceName("ingress")
			Expect(service.Name).To(Equal(expectedName))
			labels := service.Labels
			Expect(labels["app.kubernetes.io/name"]).To(Equal(instance.Name))
			Expect(labels["app.kubernetes.io/component"]).To(Equal("rabbitmq"))
			Expect(labels["app.kubernetes.io/part-of"]).To(Equal("pivotal-rabbitmq"))
		})

		By("generates a service object with the correct namespace", func() {
			Expect(service.Namespace).To(Equal(instance.Namespace))
		})

		By("generates a ClusterIP type service by default", func() {
			Expect(service.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
		})

		By("generates a service object with the correct selector", func() {
			Expect(service.Spec.Selector["app.kubernetes.io/name"]).To(Equal(instance.Name))
		})

		By("generates a service object with the correct ports exposed", func() {
			amqpPort := corev1.ServicePort{
				Name:     "amqp",
				Port:     5672,
				Protocol: corev1.ProtocolTCP,
			}
			httpPort := corev1.ServicePort{
				Name:     "http",
				Port:     15672,
				Protocol: corev1.ProtocolTCP,
			}
			prometheusPort := corev1.ServicePort{
				Name:     "prometheus",
				Port:     15692,
				Protocol: corev1.ProtocolTCP,
			}
			Expect(service.Spec.Ports).Should(ConsistOf(amqpPort, httpPort, prometheusPort))
		})

		By("generates the service without any annotation", func() {
			Expect(service.ObjectMeta.Annotations).To(BeEmpty())
		})

		By("setting the ownerreference", func() {
			Expect(service.ObjectMeta.OwnerReferences[0].Name).To(Equal("foo"))
		})
	})

	When("service type is specified in the RabbitmqCluster spec", func() {
		It("generates a service object of type LoadBalancer", func() {
			loadBalancerInstance := &rabbitmqv1beta1.RabbitmqCluster{
				ObjectMeta: v1.ObjectMeta{
					Name:      "name",
					Namespace: "mynamespace",
				},
				Spec: rabbitmqv1beta1.RabbitmqClusterSpec{
					Service: rabbitmqv1beta1.RabbitmqClusterServiceSpec{
						Type: "LoadBalancer",
					},
				},
			}
			rmqBuilder.Instance = loadBalancerInstance
			serviceBuilder := rmqBuilder.IngressService()
			obj, err := serviceBuilder.Build()
			Expect(err).NotTo(HaveOccurred())
			loadBalancerService := obj.(*corev1.Service)
			Expect(err).NotTo(HaveOccurred())
			Expect(loadBalancerService.Spec.Type).To(Equal(corev1.ServiceTypeLoadBalancer))
		})

		It("generates a service object of type ClusterIP", func() {
			clusterIPInstance := &rabbitmqv1beta1.RabbitmqCluster{
				ObjectMeta: v1.ObjectMeta{
					Name:      "name",
					Namespace: "mynamespace",
				},
				Spec: rabbitmqv1beta1.RabbitmqClusterSpec{
					Service: rabbitmqv1beta1.RabbitmqClusterServiceSpec{
						Type: "ClusterIP",
					},
				},
			}
			rmqBuilder.Instance = clusterIPInstance
			serviceBuilder := rmqBuilder.IngressService()
			obj, err := serviceBuilder.Build()
			Expect(err).NotTo(HaveOccurred())
			clusterIPService := obj.(*corev1.Service)
			Expect(clusterIPService.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
		})

		It("generates a service object of type NodePort", func() {
			nodePortInstance := &rabbitmqv1beta1.RabbitmqCluster{
				ObjectMeta: v1.ObjectMeta{
					Name:      "name",
					Namespace: "mynamespace",
				},
				Spec: rabbitmqv1beta1.RabbitmqClusterSpec{
					Service: rabbitmqv1beta1.RabbitmqClusterServiceSpec{
						Type: "NodePort",
					},
				},
			}
			rmqBuilder.Instance = nodePortInstance
			serviceBuilder := rmqBuilder.IngressService()
			obj, err := serviceBuilder.Build()
			Expect(err).NotTo(HaveOccurred())
			nodePortService := obj.(*corev1.Service)
			Expect(err).NotTo(HaveOccurred())
			Expect(nodePortService.Spec.Type).To(Equal(corev1.ServiceTypeNodePort))
		})
	})

	Context("Annotations", func() {
		When("service annotations are specified on the instance", func() {
			It("generates the service annotations as specified in the RabbitmqCluster spec", func() {
				serviceAnno := map[string]string{
					"service_annotation_a":       "0.0.0.0/0",
					"kubernetes.io/name":         "i-do-not-like-this",
					"kubectl.kubernetes.io/name": "i-do-not-like-this",
					"k8s.io/name":                "i-do-not-like-this",
				}
				expectedAnnotations := map[string]string{"service_annotation_a": "0.0.0.0/0"}
				service := getServiceWithAnnotations(rmqBuilder, nil, serviceAnno)
				Expect(service.ObjectMeta.Annotations).To(Equal(expectedAnnotations))
			})
		})

		When("instance annotations set on the instance, service annotations are not set on instance", func() {
			It("sets the instance annotations on the service", func() {
				instanceAnno := map[string]string{
					"my-annotation":              "i-like-this",
					"kubernetes.io/name":         "i-do-not-like-this",
					"kubectl.kubernetes.io/name": "i-do-not-like-this",
					"k8s.io/name":                "i-do-not-like-this",
				}
				service := getServiceWithAnnotations(rmqBuilder, instanceAnno, nil)
				Expect(service.ObjectMeta.Annotations).To(Equal(map[string]string{"my-annotation": "i-like-this"}))
			})
		})

		When("instance annotations set on the instance, service annotations set on instance", func() {
			It("merges the annotations", func() {
				serviceAnno := map[string]string{
					"service_annotation_a":       "0.0.0.0/0",
					"my-annotation":              "i-like-this-more",
					"kubernetes.io/name":         "i-do-not-like-this",
					"kubectl.kubernetes.io/name": "i-do-not-like-this",
					"k8s.io/name":                "i-do-not-like-this",
				}
				instanceAnno := map[string]string{
					"my-annotation":              "i-like-this",
					"my-second-annotation":       "i-like-this-also",
					"kubernetes.io/name":         "i-do-not-like-this",
					"kubectl.kubernetes.io/name": "i-do-not-like-this",
					"k8s.io/name":                "i-do-not-like-this",
				}

				service := getServiceWithAnnotations(rmqBuilder, instanceAnno, serviceAnno)
				Expect(service.ObjectMeta.Annotations).To(Equal(map[string]string{
					"my-annotation":        "i-like-this-more",
					"my-second-annotation": "i-like-this-also",
					"service_annotation_a": "0.0.0.0/0",
				},
				))
			})
		})
	})

	Context("label inheritance", func() {
		BeforeEach(func() {
			instance.Labels = map[string]string{
				"app.kubernetes.io/foo": "bar",
				"foo":                   "bar",
				"rabbitmq":              "is-great",
				"foo/app.kubernetes.io": "edgecase",
			}
		})

		It("has the labels from the CRD on the ingress service", func() {
			serviceBuilder := rmqBuilder.IngressService()
			obj, err := serviceBuilder.Build()
			Expect(err).NotTo(HaveOccurred())
			ingressService := obj.(*corev1.Service)
			testLabels(ingressService.Labels)
		})
	})

	Context("Update", func() {
		Context("Annotations", func() {
			When("CR instance does have service annotations specified", func() {
				It("generates a service object with the annotations as specified", func() {
					serviceAnno := map[string]string{
						"service_annotation_a":       "0.0.0.0/0",
						"kubernetes.io/name":         "i-do-not-like-this",
						"kubectl.kubernetes.io/name": "i-do-not-like-this",
						"k8s.io/name":                "i-do-not-like-this",
					}
					expectedAnnotations := map[string]string{
						"service_annotation_a":             "0.0.0.0/0",
						"app.kubernetes.io/part-of":        "pivotal-rabbitmq",
						"app.k8s.io/something":             "something-amazing",
						"this-was-the-previous-annotation": "should-be-preserved",
					}

					service := updateServiceWithAnnotations(rmqBuilder, nil, serviceAnno)
					Expect(service.ObjectMeta.Annotations).To(Equal(expectedAnnotations))
				})
			})

			When("CR instance does not have service annotations specified", func() {
				It("generates the service annotations as specified", func() {
					expectedAnnotations := map[string]string{
						"app.kubernetes.io/part-of":        "pivotal-rabbitmq",
						"app.k8s.io/something":             "something-amazing",
						"this-was-the-previous-annotation": "should-be-preserved",
					}

					var serviceAnnotations map[string]string = nil
					var instanceAnnotations map[string]string = nil
					service := updateServiceWithAnnotations(rmqBuilder, instanceAnnotations, serviceAnnotations)
					Expect(service.ObjectMeta.Annotations).To(Equal(expectedAnnotations))
				})
			})

			When("CR instance does not have service annotations specified, but does have metadata annotations specified", func() {
				It("sets the instance annotations on the service", func() {
					instanceMetadataAnnotations := map[string]string{
						"my-annotation":              "i-like-this",
						"kubernetes.io/name":         "i-do-not-like-this",
						"kubectl.kubernetes.io/name": "i-do-not-like-this",
						"k8s.io/name":                "i-do-not-like-this",
					}

					var serviceAnnotations map[string]string = nil
					service := updateServiceWithAnnotations(rmqBuilder, instanceMetadataAnnotations, serviceAnnotations)
					expectedAnnotations := map[string]string{
						"my-annotation":                    "i-like-this",
						"app.kubernetes.io/part-of":        "pivotal-rabbitmq",
						"app.k8s.io/something":             "something-amazing",
						"this-was-the-previous-annotation": "should-be-preserved",
					}

					Expect(service.Annotations).To(Equal(expectedAnnotations))
				})
			})

			When("CR instance has service annotations specified, and has metadata annotations specified", func() {
				It("merges the annotations", func() {
					serviceAnnotations := map[string]string{
						"service_annotation_a":       "0.0.0.0/0",
						"my-annotation":              "i-like-this-more",
						"kubernetes.io/name":         "i-do-not-like-this",
						"kubectl.kubernetes.io/name": "i-do-not-like-this",
						"k8s.io/name":                "i-do-not-like-this",
					}
					instanceAnnotations := map[string]string{
						"my-annotation":              "i-like-this",
						"my-second-annotation":       "i-like-this-also",
						"kubernetes.io/name":         "i-do-not-like-this",
						"kubectl.kubernetes.io/name": "i-do-not-like-this",
						"k8s.io/name":                "i-do-not-like-this",
					}

					expectedAnnotations := map[string]string{
						"my-annotation":                    "i-like-this-more",
						"my-second-annotation":             "i-like-this-also",
						"service_annotation_a":             "0.0.0.0/0",
						"app.kubernetes.io/part-of":        "pivotal-rabbitmq",
						"app.k8s.io/something":             "something-amazing",
						"this-was-the-previous-annotation": "should-be-preserved",
					}

					service := updateServiceWithAnnotations(rmqBuilder, instanceAnnotations, serviceAnnotations)

					Expect(service.ObjectMeta.Annotations).To(Equal(expectedAnnotations))
				})
			})
		})

		Context("Labels", func() {
			var (
				serviceBuilder *resource.IngressServiceBuilder
				ingressService *corev1.Service
			)
			BeforeEach(func() {
				serviceBuilder = rmqBuilder.IngressService()
				instance = rabbitmqv1beta1.RabbitmqCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "rabbit-labelled",
					},
				}
				instance.Labels = map[string]string{
					"app.kubernetes.io/foo": "bar",
					"foo":                   "bar",
					"rabbitmq":              "is-great",
					"foo/app.kubernetes.io": "edgecase",
				}

				ingressService = &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app.kubernetes.io/name":      instance.Name,
							"app.kubernetes.io/part-of":   "pivotal-rabbitmq",
							"this-was-the-previous-label": "should-be-deleted",
						},
					},
				}
				err := serviceBuilder.Update(ingressService)
				Expect(err).NotTo(HaveOccurred())
			})

			It("adds labels from the CR", func() {
				testLabels(ingressService.Labels)
			})

			It("restores the default labels", func() {
				labels := ingressService.Labels
				Expect(labels["app.kubernetes.io/name"]).To(Equal(instance.Name))
				Expect(labels["app.kubernetes.io/component"]).To(Equal("rabbitmq"))
				Expect(labels["app.kubernetes.io/part-of"]).To(Equal("pivotal-rabbitmq"))
			})

			It("deletes the labels that are removed from the CR", func() {
				Expect(ingressService.Labels).NotTo(HaveKey("this-was-the-previous-label"))
			})
		})

		Context("Service Type", func() {
			var (
				ingressService *corev1.Service
			)

			BeforeEach(func() {
				serviceBuilder := rmqBuilder.IngressService()
				instance = rabbitmqv1beta1.RabbitmqCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "rabbit-service-type-update",
					},
					Spec: rabbitmqv1beta1.RabbitmqClusterSpec{
						Service: rabbitmqv1beta1.RabbitmqClusterServiceSpec{
							Type: "NodePort",
						},
					},
				}

				ingressService = &corev1.Service{
					Spec: corev1.ServiceSpec{
						Type: corev1.ServiceTypeClusterIP,
					},
				}
				err := serviceBuilder.Update(ingressService)
				Expect(err).NotTo(HaveOccurred())
			})

			It("updates the service type", func() {
				expectedServiceType := "NodePort"
				Expect(string(ingressService.Spec.Type)).To(Equal(expectedServiceType))
			})
		})
	})
})

func getServiceWithAnnotations(rmqBuilder resource.RabbitmqResourceBuilder, instanceAnno, serviceAnno map[string]string) *corev1.Service {
	instance := &rabbitmqv1beta1.RabbitmqCluster{
		ObjectMeta: v1.ObjectMeta{
			Name:        "name",
			Namespace:   "mynamespace",
			Annotations: instanceAnno,
		},
		Spec: rabbitmqv1beta1.RabbitmqClusterSpec{
			Service: rabbitmqv1beta1.RabbitmqClusterServiceSpec{
				Annotations: serviceAnno,
			},
		},
	}

	rmqBuilder.Instance = instance
	serviceBuilder := rmqBuilder.IngressService()
	obj, err := serviceBuilder.Build()
	Expect(err).NotTo(HaveOccurred())
	service := obj.(*corev1.Service)
	Expect(err).NotTo(HaveOccurred())
	return service
}

func updateServiceWithAnnotations(rmqBuilder resource.RabbitmqResourceBuilder, instanceAnnotations, serviceAnnotations map[string]string) *corev1.Service {
	instance := &rabbitmqv1beta1.RabbitmqCluster{
		ObjectMeta: v1.ObjectMeta{
			Name:        "name",
			Namespace:   "mynamespace",
			Annotations: instanceAnnotations,
		},
		Spec: rabbitmqv1beta1.RabbitmqClusterSpec{
			Service: rabbitmqv1beta1.RabbitmqClusterServiceSpec{
				Annotations: serviceAnnotations,
			},
		},
	}

	rmqBuilder.Instance = instance
	serviceBuilder := rmqBuilder.IngressService()
	ingressService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"this-was-the-previous-annotation": "should-be-preserved",
				"app.kubernetes.io/part-of":        "pivotal-rabbitmq",
				"app.k8s.io/something":             "something-amazing",
			},
		},
	}
	Expect(serviceBuilder.Update(ingressService)).To(Succeed())
	return ingressService
}
