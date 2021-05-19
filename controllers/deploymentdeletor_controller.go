/*
Copyright 2021.

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

package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	mikutasv1alpha1 "github.com/mikutas/deployment-deletor/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DeploymentDeletorReconciler reconciles a DeploymentDeletor object
type DeploymentDeletorReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=mikutas.example.com,resources=deploymentdeletors,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=mikutas.example.com,resources=deploymentdeletors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mikutas.example.com,resources=deploymentdeletors/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=list;watch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DeploymentDeletor object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *DeploymentDeletorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("DeploymentDeletor", req.NamespacedName)

	// your logic here

	var dd mikutasv1alpha1.DeploymentDeletor
	log.Info("Fetching DeploymentDeletor Resource")
	if err := r.Get(ctx, req.NamespacedName, &dd); err != nil {
		log.Error(err, "Inable to fetch DeploymentDeletor")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	duration, err := time.ParseDuration(dd.Spec.MaxAge)
	if err != nil {
		log.Error(err, "Invalid maxAge field")
		return ctrl.Result{}, err
	}

	labelSelector, err := metav1.LabelSelectorAsSelector(dd.Spec.Selector)
	if err != nil {
		log.Error(err, "unable to parse selector")
		return ctrl.Result{}, err
	}
	log.Info("", "labelSelector", labelSelector)

	var deployments appsv1.DeploymentList
	if err := r.List(ctx, &deployments, &client.ListOptions{
		LabelSelector: labelSelector,
		Namespace:     dd.Spec.Deployment.Namespace,
	}); err != nil {
		log.Error(err, "unable to fetch deployments")
	}
	log.Info("", "deployments", deployments)

	var lastDeleted *appsv1.Deployment
	for _, deployment := range deployments.Items {
		if dd.Spec.Deployment.Name != "" && dd.Spec.Deployment.Name != deployment.Name {
			log.Info("Name not matched", "dd.Spec.Deployment.Name", dd.Spec.Deployment.Name, "deployment.Name", deployment.Name)
			continue
		}
		if exceedsMaxAge(&deployment, duration) {
			log.Info("Deleting Deployment", "namespace", deployment.Namespace, "name", deployment.Name)
			if err := r.Delete(ctx, &deployment); err != nil {
				log.Error(err, "unable to delete Deployment")
				r.Recorder.Eventf(&dd, v1.EventTypeNormal, "FailedDeleting", "Failed to delete Deployment %q", deployment.Name)
				continue
			}
			log.Info("Deleted Deployment", "namespace", deployment.Namespace, "name", deployment.Name)
			r.Recorder.Eventf(&dd, v1.EventTypeNormal, "Deleted", "Deleted Deployment %q", deployment.Name)
			lastDeleted = deployment.DeepCopy()
		} else {
			log.Info("This Deployment needs not to delete", "namespace", deployment.Namespace, "name", deployment.Name)
		}
	}

	if lastDeleted == nil {
		log.Info("Nothing is deleted in this reconciliation")
		return ctrl.Result{}, nil
	}

	dd.Status.LastDeletedDeployment = *lastDeleted.ObjectMeta.DeepCopy()
	if err := r.Status().Update(ctx, &dd); err != nil {
		log.Error(err, "Unable to update deploymentdeletor status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentDeletorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mikutasv1alpha1.DeploymentDeletor{}).
		Complete(r)
}

func exceedsMaxAge(deployment *appsv1.Deployment, duration time.Duration) bool {
	return deployment.CreationTimestamp.Add(duration).Before(time.Now().UTC())
}
