/*
Copyright 2026.

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

package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	log "sigs.k8s.io/controller-runtime/pkg/log"

	apiv1alpha1 "github.com/Toandos/stalwart-operator/api/v1alpha1"
)

const (
	portNameSMTP       = "smtp"
	portNameSMTPS      = "smtps"
	portNameSubmission = "submission"
	portNameIMAP       = "imap"
	portNameIMAPS      = "imaps"
	portNamePOP3       = "pop3"
	portNamePOP3S      = "pop3s"
	portNameSieve      = "sieve"
	portNameHTTP       = "http"
	portNameHTTPS      = "https"
	portNameMgmt       = "mgmt"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	cluster, err := r.getCluster(ctx, req)
	if err != nil {
		return ctrl.Result{}, err
	}
	if cluster == nil {
		return ctrl.Result{}, nil
	}

	if err := r.reconcileConfigMap(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.reconcileDeployment(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.reconcileHeadlessService(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.reconcileService(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) getCluster(
	ctx context.Context,
	req ctrl.Request,
) (*apiv1alpha1.Cluster, error) {
	logger := log.FromContext(ctx)
	cluster := &apiv1alpha1.Cluster{}
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		if apierrs.IsNotFound(err) {
			logger.Info("Resource has been deleted")
			return nil, nil
		}
		return nil, fmt.Errorf("cannot get managed cluster resource: %w", err)
	}
	return cluster, nil
}

func getLabels(cluster *apiv1alpha1.Cluster) map[string]string {
	labels := map[string]string{
		"app.kubernetes.io/name":       "stalwart",
		"app.kubernetes.io/instance":   cluster.Name,
		"app.kubernetes.io/managed-by": "cluster-operator",
	}
	return labels
}

func (r *ClusterReconciler) reconcileConfigMap(ctx context.Context, cluster *apiv1alpha1.Cluster) error {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-config",
			Namespace: cluster.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, configMap, func() error {
		configMap.Data = map[string]string{
			"config.json": `{
				"@type": "PostgreSql",
				"host":  "postgres.postgres",
			}`,
		}

		// Make the Cluster the owner of the ConfigMap.
		return controllerutil.SetControllerReference(
			cluster,
			configMap,
			r.Scheme,
		)
	})

	return err
}

func (r *ClusterReconciler) reconcileDeployment(ctx context.Context, cluster *apiv1alpha1.Cluster) error {
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, statefulSet, func() error {
		statefulSet.Spec.Replicas = ptr.To(int32(cluster.Spec.Instances))

		statefulSet.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: getLabels(cluster),
		}

		statefulSet.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: getLabels(cluster),
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "stalwart",
						Image: "stalwartlabs/stalwart:latest",
						Ports: []corev1.ContainerPort{
							{
								Name:          portNameSMTP,
								ContainerPort: 25,
							},
							{
								Name:          portNameSMTPS,
								ContainerPort: 465,
							},
							{
								Name:          portNameSubmission,
								ContainerPort: 587,
							},
							{
								Name:          portNameIMAP,
								ContainerPort: 143,
							},
							{
								Name:          portNameIMAPS,
								ContainerPort: 993,
							},
							{
								Name:          portNamePOP3,
								ContainerPort: 110,
							},
							{
								Name:          portNamePOP3S,
								ContainerPort: 995,
							},
							{
								Name:          portNameSieve,
								ContainerPort: 4190,
							},
							{
								Name:          portNameHTTP,
								ContainerPort: 80,
							},
							{
								Name:          portNameHTTPS,
								ContainerPort: 443,
							},
							{
								Name:          portNameMgmt,
								ContainerPort: 8080,
							},
						},
						Args: []string{
							"--config",
							"/etc/stalwart/config.json",
						},
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/healthz/live",
									Port: intstr.FromString(portNameMgmt),
								},
							},
							InitialDelaySeconds: 5,
							PeriodSeconds:       10,
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/healthz/ready",
									Port: intstr.FromString(portNameMgmt),
								},
							},
							InitialDelaySeconds: 30,
							PeriodSeconds:       10,
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "config",
								MountPath: "/etc/stalwart/config.json",
								SubPath:   "config.json",
								ReadOnly:  true,
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "config",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: statefulSet.Name + "-config",
								},
							},
						},
					},
				},
			},
		}

		// Make the Cluster the owner of the Deployment.
		return controllerutil.SetControllerReference(
			cluster,
			statefulSet,
			r.Scheme,
		)
	})

	return err
}

func (r *ClusterReconciler) reconcileHeadlessService(ctx context.Context, cluster *apiv1alpha1.Cluster) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-headless",
			Namespace: cluster.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		service.Spec.ClusterIP = "None"
		service.Spec.Selector = getLabels(cluster)
		service.Spec.Ports = []corev1.ServicePort{
			{
				Name:       portNameMgmt,
				Port:       8080,
				TargetPort: intstr.FromString(portNameMgmt),
			},
		}

		// Make the Cluster the owner of the Service.
		return controllerutil.SetControllerReference(
			cluster,
			service,
			r.Scheme,
		)
	})

	return err
}

func (r *ClusterReconciler) reconcileService(ctx context.Context, cluster *apiv1alpha1.Cluster) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		service.Spec.Selector = getLabels(cluster)
		service.Spec.Ports = []corev1.ServicePort{
			{
				Name:       portNameSMTP,
				Port:       25,
				TargetPort: intstr.FromString(portNameSMTP),
			},
			{
				Name:       portNameSMTPS,
				Port:       465,
				TargetPort: intstr.FromString(portNameSMTPS),
			},
			{
				Name:       portNameSubmission,
				Port:       587,
				TargetPort: intstr.FromString(portNameSubmission),
			},
			{
				Name:       portNameIMAP,
				Port:       143,
				TargetPort: intstr.FromString(portNameIMAP),
			},
			{
				Name:       portNameIMAPS,
				Port:       993,
				TargetPort: intstr.FromString(portNameIMAPS),
			},
			{
				Name:       portNamePOP3,
				Port:       110,
				TargetPort: intstr.FromString(portNamePOP3),
			},
			{
				Name:       portNamePOP3S,
				Port:       995,
				TargetPort: intstr.FromString(portNamePOP3S),
			},
			{
				Name:       portNameSieve,
				Port:       4190,
				TargetPort: intstr.FromString(portNameSieve),
			},
			{
				Name:       portNameHTTP,
				Port:       80,
				TargetPort: intstr.FromString(portNameHTTP),
			},
			{
				Name:       portNameHTTPS,
				Port:       443,
				TargetPort: intstr.FromString(portNameHTTPS),
			},
			{
				Name:       portNameMgmt,
				Port:       8080,
				TargetPort: intstr.FromString(portNameMgmt),
			},
		}

		// Make the Cluster the owner of the Service.
		return controllerutil.SetControllerReference(
			cluster,
			service,
			r.Scheme,
		)
	})

	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.Cluster{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
