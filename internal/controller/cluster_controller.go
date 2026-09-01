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
	"k8s.io/utils/ptr"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	log "sigs.k8s.io/controller-runtime/pkg/log"

	apiv1alpha1 "github.com/Toandos/stalwart-operator/api/v1alpha1"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=stalwart.toando.de,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=stalwart.toando.de,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=stalwart.toando.de,resources=clusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	cluster, err := r.getCluster(ctx, req)
	if err != nil {
		return ctrl.Result{}, err
	}
	if cluster == nil {
		return ctrl.Result{}, nil
	}

	if err := r.reconcileDeployment(ctx, cluster); err != nil {
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

func (r *ClusterReconciler) reconcileDeployment(ctx context.Context, cluster *apiv1alpha1.Cluster) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		labels := map[string]string{
			"app.kubernetes.io/name":       "stalwart",
			"app.kubernetes.io/instance":   cluster.Name,
			"app.kubernetes.io/managed-by": "cluster-operator",
		}

		deployment.Spec.Replicas = ptr.To(int32(1))

		deployment.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: labels,
		}

		deployment.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "stalwart",
						Image: "stalwartlabs/stalwart:latest",
						Ports: []corev1.ContainerPort{
							{
								Name:          "http",
								ContainerPort: 8080,
							},
							{
								Name:          "smtp",
								ContainerPort: 25,
							},
						},
					},
				},
			},
		}

		// Make the Cluster the owner of the Deployment.
		return controllerutil.SetControllerReference(
			cluster,
			deployment,
			r.Scheme,
		)
	})

	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.Cluster{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
