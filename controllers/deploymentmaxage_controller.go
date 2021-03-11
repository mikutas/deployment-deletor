/*


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
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tkms0106v1alpha1 "github.com/tkms0106/deployment-deletor/api/v1alpha1"
)

// DeploymentMaxAgeReconciler reconciles a DeploymentMaxAge object
type DeploymentMaxAgeReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=tkms0106.example.com,resources=deploymentmaxages,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tkms0106.example.com,resources=deploymentmaxages/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=list
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create,patch

func (r *DeploymentMaxAgeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("deploymentmaxage", req.NamespacedName)

	// your logic here

	var maxage tkms0106v1alpha1.DeploymentMaxAge
	log.Info("fetching DeploymentMaxAge Resource")
	if err := r.Get(ctx, req.NamespacedName, &maxage); err != nil {
		log.Error(err, "unable to fetch DeploymentMaxAge")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	duration, err := time.ParseDuration(maxage.Spec.MaxAge)
	if err != nil {
		log.Error(err, "invalid MaxAge field")
		return ctrl.Result{}, err
	}

	var namespaces v1.NamespaceList
	if err := r.List(ctx, &namespaces); err != nil {
		log.Error(err, "unable to list Namespaces")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	lastDeleted := &appsv1.Deployment{}
	for _, namespace := range namespaces.Items {
		var deployment appsv1.Deployment
		target := types.NamespacedName{
			Namespace: namespace.Name,
			Name:      maxage.Spec.DeploymentName,
		}
		if err := r.Get(ctx, target, &deployment); err != nil {
			log.Error(err, "unable to fetch Deployment in %q", namespace.Name)
		}
		if exceedsMaxAge(&deployment, duration) {
			log.Info("deleting Deployment: %s/%s", deployment.Namespace, deployment.Name)
			if err := r.Delete(ctx, &deployment); err != nil {
				log.Error(err, "unable to delete Deployment: %s/%s", deployment.Namespace, deployment.Name)
				r.Recorder.Eventf(&maxage, v1.EventTypeNormal, "FailedDeleting", "Failed to delete deployment %q", deployment.Name)
			}
			log.Info("deleted Deployment: %s/%s", deployment.Namespace, deployment.Name)
			r.Recorder.Eventf(&maxage, v1.EventTypeNormal, "Deleted", "Deleted deployment %q", deployment.Name)
			lastDeleted = deployment.DeepCopy()
		}
	}

	maxage.Status.LastDeletedDeployment = *lastDeleted.ObjectMeta.DeepCopy()
	if err := r.Status().Update(ctx, &maxage); err != nil {
		log.Error(err, "unable to update maxage status")
		return ctrl.Result{}, err
	}
	r.Recorder.Eventf(&maxage, v1.EventTypeNormal, "Updated", "Update maxage.status.LastDeletedDeployment: %q", &maxage.Status.LastDeletedDeployment.Name)

	return ctrl.Result{}, nil
}

func (r *DeploymentMaxAgeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tkms0106v1alpha1.DeploymentMaxAge{}).
		Complete(r)
}

func exceedsMaxAge(deployment *appsv1.Deployment, duration time.Duration) bool {
	return deployment.CreationTimestamp.Add(duration).After(time.Now().UTC())
}
