/*
Copyright 2022.

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

package alerts

import (
	"context"

	e "errors"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	alertsv1alpha1 "github.com/VioletCranberry/scalyr-operator-test/apis/alerts/v1alpha1"

	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const alertRuleFinalizer = "scalyralertrule.scalyr.com/finalizer"

// ScalyrAlertRuleReconciler reconciles a ScalyrAlertRule object
type ScalyrAlertRuleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *ScalyrAlertRuleReconciler) validateAlertRule(ctx context.Context,
	resourceList *alertsv1alpha1.ScalyrAlertRuleList,
	resource *alertsv1alpha1.ScalyrAlertRule) error {
	specTriggerExists := false
	for _, item := range resourceList.Items {
		if item.Spec.Trigger == resource.Spec.Trigger {
			if !specTriggerExists {
				specTriggerExists = true
				continue
			} else {
				return e.New("Similar trigger already exists. " +
					"Resource field spec.Trigger must be unique")
			}
		}
	}
	return nil
}

func (r *ScalyrAlertRuleReconciler) finalizeAlertRule(ctx context.Context,
	alertFileResource *alertsv1alpha1.ScalyrAlertFile,
	alertRuleResource *alertsv1alpha1.ScalyrAlertRule) error {
	for i, alertSpec := range alertFileResource.Spec.Alerts {
		if alertSpec.Trigger == alertRuleResource.Spec.Trigger {
			alertFileResource.Spec.Alerts = append(
				alertFileResource.Spec.Alerts[:i],
				alertFileResource.Spec.Alerts[i+1:]...)
		}
	}
	return r.Update(ctx, alertFileResource)
}

func (r *ScalyrAlertRuleReconciler) updateAlertFile(ctx context.Context,
	alertFileResource *alertsv1alpha1.ScalyrAlertFile,
	alertRuleResource *alertsv1alpha1.ScalyrAlertRule) error {
	log := log.FromContext(ctx)
	specTriggerExists := false
	for _, alertSpec := range alertFileResource.Spec.Alerts {
		if alertSpec.Trigger == alertRuleResource.Spec.Trigger {
			specTriggerExists = true
			alertSpec = alertRuleResource.Spec
		}
	}
	if !specTriggerExists {
		alertFileResource.Spec.Alerts = append(
			alertFileResource.Spec.Alerts,
			alertRuleResource.Spec)
	}
	log.Info("ScalyrAlertFile resource updated",
		"ScalyrAlertFile.Name",
		alertFileResource.Name,
		"ScalyrAlertFile.Namespace",
		alertFileResource.Namespace)
	return r.Update(ctx, alertFileResource)
}

//+kubebuilder:rbac:groups=alerts.scalyr.com,resources=scalyralertrules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=alerts.scalyr.com,resources=scalyralertrules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=alerts.scalyr.com,resources=scalyralertrules/finalizers,verbs=update

func (r *ScalyrAlertRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	var alertRule alertsv1alpha1.ScalyrAlertRule
	var alertRuleList alertsv1alpha1.ScalyrAlertRuleList
	var alertFileList alertsv1alpha1.ScalyrAlertFileList

	if err := r.Get(ctx, req.NamespacedName, &alertRule); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Resource not found. " +
				"Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if err := r.List(ctx, &alertRuleList, &client.ListOptions{
		Namespace: req.Namespace}); err != nil {
		log.Error(err, "Failed to list resources")
		return ctrl.Result{}, err
	}
	if err := r.List(ctx, &alertFileList, &client.ListOptions{
		Namespace: req.Namespace}); err != nil {
		log.Error(err, "Failed to list ScalyrAlertFile resources")
		return ctrl.Result{}, err
	}
	if len(alertFileList.Items) == 0 {
		log.Info("ScalyrAlertFile resources not found. " +
			"Skipping the rest of reconciliation loop")
		return ctrl.Result{Requeue: true}, nil
	}

	isResourceToBeDeleted := alertRule.GetDeletionTimestamp() != nil
	if isResourceToBeDeleted {
		for _, resource := range alertFileList.Items {
			if err := r.Get(ctx, client.ObjectKey{
				Namespace: req.Namespace,
				Name:      resource.Name},
				&resource); err != nil {
				log.Error(err, "Resource scalyrAlertFile not found")
				return ctrl.Result{}, err
			}
			if controllerutil.ContainsFinalizer(&alertRule, alertRuleFinalizer) {
				if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
					return r.finalizeAlertRule(ctx, &resource, &alertRule)
				}); err != nil {
					return ctrl.Result{}, err
				}
			}
		}
		controllerutil.RemoveFinalizer(&alertRule, alertRuleFinalizer)
		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return r.Update(ctx, &alertRule)
		}); err != nil {
			return ctrl.Result{}, err
		}
	}
	if !isResourceToBeDeleted {
		if err := r.validateAlertRule(ctx, &alertRuleList, &alertRule); err != nil {
			log.Error(err, "Failed resource falidation")
		}
		for _, resource := range alertFileList.Items {
			if err := r.Get(ctx, client.ObjectKey{
				Namespace: req.Namespace,
				Name:      resource.Name},
				&resource); err != nil {
				log.Error(err, "Resource scalyrAlertFile not found")
				return ctrl.Result{}, err
			}

			if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				return r.updateAlertFile(ctx, &resource, &alertRule)
			}); err != nil {
				return ctrl.Result{}, err
			}
		}
		if !controllerutil.ContainsFinalizer(&alertRule, alertRuleFinalizer) {
			controllerutil.AddFinalizer(
				&alertRule, alertRuleFinalizer)
			if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				return r.Update(ctx, &alertRule)
			}); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	alertRule.Status.Synchronised = true
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return r.Status().Update(ctx, &alertRule)
	}); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Resource reconciled")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScalyrAlertRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&alertsv1alpha1.ScalyrAlertRule{}).
		WithOptions(controller.Options{
			RecoverPanic: true}).
		Complete(r)
}
