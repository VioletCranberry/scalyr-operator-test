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
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	e "errors"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	alertsv1alpha1 "github.com/VioletCranberry/scalyr-operator-test/apis/alerts/v1alpha1"
	scalyr "github.com/VioletCranberry/scalyr-operator-test/internal/clients/scalyr"
)

const alertFileFinalizer = "scalyralertfile.scalyr.com/finalizer"

// ScalyrAlertFileReconciler reconciles a ScalyrAlertFile object
type ScalyrAlertFileReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	ScalyrApiClient *scalyr.ApiClient
}

func (r *ScalyrAlertFileReconciler) validateAlertFile(ctx context.Context,
	resourceList *alertsv1alpha1.ScalyrAlertFileList) error {
	log := log.FromContext(ctx)
	if len(resourceList.Items) > 1 {
		return e.New("Resource of the same type already exists. " +
			"Only one resource of such type is allowed")
	}
	log.Info("Resource validated")
	return nil
}

func (r *ScalyrAlertFileReconciler) finalizeAlertFile(ctx context.Context,
	scalyrConfigFilePath string) error {
	log := log.FromContext(ctx)
	// /scalyr/alerts cannot be deleted and empty content {}
	//  blocks alerts creation via UI. Set is to specific one
	jsonBody := `{"alertAddress": "","alerts": []}`
	response, err := r.ScalyrApiClient.PutFile(scalyrConfigFilePath, jsonBody)
	if err != nil {
		return err
	}
	if response.Status != "success" {
		return fmt.Errorf(
			"Invalid API response! message: %s",
			response.Message)
	}
	log.Info("Resource finalized")
	return nil
}

func (r *ScalyrAlertFileReconciler) updateScalyrConfigFile(ctx context.Context,
	scalyrConfigFilePath string,
	resource *alertsv1alpha1.ScalyrAlertFile) error {
	log := log.FromContext(ctx)

	jsonBuf := new(bytes.Buffer)
	jsonEnc := json.NewEncoder(jsonBuf)
	jsonEnc.SetEscapeHTML(false)

	err := jsonEnc.Encode(resource.Spec)
	if err != nil {
		return fmt.Errorf("Error occured during marshaling! error: %s", err)
	}
	response, err := r.ScalyrApiClient.PutFile(
		scalyrConfigFilePath, string(jsonBuf.String()))
	if err != nil {
		return err
	}
	if response.Status != "success" {
		return fmt.Errorf(
			"Invalid API response! message: %s",
			response.Message)
	}
	log.Info("Scalyr Config File updated")
	return nil
}

//+kubebuilder:rbac:groups=alerts.scalyr.com,resources=scalyralertfiles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=alerts.scalyr.com,resources=scalyralertfiles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=alerts.scalyr.com,resources=scalyralertfiles/finalizers,verbs=update

func (r *ScalyrAlertFileReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	var alertFile alertsv1alpha1.ScalyrAlertFile
	var alertFileList alertsv1alpha1.ScalyrAlertFileList
	var alertRuleList alertsv1alpha1.ScalyrAlertRuleList

	scalyrConfigFilePath := "/scalyr/alerts"

	if err := r.Get(ctx, req.NamespacedName, &alertFile); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Resource not found. " +
				"Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if err := r.List(ctx, &alertFileList, &client.ListOptions{
		Namespace: req.Namespace}); err != nil {
		log.Error(err, "Failed to list resources")
		return ctrl.Result{}, err
	}
	if err := r.validateAlertFile(ctx, &alertFileList); err != nil {
		log.Error(err, "Failed resource falidation")
	}
	isResourceToBeDeleted := alertFile.GetDeletionTimestamp() != nil
	if isResourceToBeDeleted {
		if controllerutil.ContainsFinalizer(&alertFile, alertFileFinalizer) {
			if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				return r.finalizeAlertFile(ctx, scalyrConfigFilePath)
			}); err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(&alertFile, alertFileFinalizer)
			if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				return r.Update(ctx, &alertFile)
			}); err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	if !isResourceToBeDeleted {
		for _, resource := range alertRuleList.Items {
			if err := r.Get(ctx, client.ObjectKey{
				Namespace: req.Namespace,
				Name:      resource.Name},
				&resource); err != nil {
				log.Error(err, "Resource scalyrAlertRule not found")
				return ctrl.Result{}, err
			}
			if resource.Status.Synchronised != true {
				log.Info("scalyrAlertRule resource is not in sync yet",
					resource.Name,
					resource.Namespace)
				return ctrl.Result{Requeue: true}, nil
			}
		}
		err := r.updateScalyrConfigFile(ctx, scalyrConfigFilePath, &alertFile)
		if err != nil {
			return ctrl.Result{}, err
		}
		if !controllerutil.ContainsFinalizer(&alertFile, alertFileFinalizer) {
			controllerutil.AddFinalizer(
				&alertFile, alertFileFinalizer)
			if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				return r.Update(ctx, &alertFile)
			}); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	alertFile.Status.Synchronised = true
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return r.Status().Update(ctx, &alertFile)
	}); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Resource reconciled")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScalyrAlertFileReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&alertsv1alpha1.ScalyrAlertFile{}).
		WithOptions(controller.Options{
			RecoverPanic: true}).
		Complete(r)
}
