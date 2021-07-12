package controllers

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// PodReconcilerUpdates reconciles a Pod object for pod updates only (ie not create/delete events)
type PodReconcilerUpdates struct {
	client.Client
	Scheme            *runtime.Scheme
	Slack             SlackConnector
	PodTargetContains string
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=get

// Reconcile reconciles changes to pod resources when they are updated
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *PodReconcilerUpdates) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx).WithValues("pod", req.NamespacedName)
	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		// pod has been deleted
		if apierrors.IsNotFound(err) {
			l.Info("UPDATECONTROLLER - should never fire!!!", "deletedName", req.NamespacedName.Name)
			// l.Info("UPDATECONTROLLER", "deleted", pod)
			return ctrl.Result{}, nil
		}
		l.Error(err, "UPDATECONTROLLER - should never fire!!! unable to fetch Pod")
		return ctrl.Result{}, err
	}
	// l.Info("UPDATECONTROLLER", "status", pod.Status)
	l.Info("UPDATECONTROLLER", "name", pod.Name)
	err := r.Slack.PostMessage(fmt.Sprintf("things have changed, %s", pod.Name))
	return ctrl.Result{}, err
}

func ignoreEverythingButUpdatesPredicate(podTargetContains string) predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// // Only consider updates to pods containing our target pod name
			if strings.Contains(e.ObjectNew.GetName(), podTargetContains) {
				if e.ObjectOld == nil {
					return false
				}
				if e.ObjectNew == nil {
					return false
				}
				return !reflect.DeepEqual(e.ObjectNew.GetLabels(), e.ObjectOld.GetLabels())
			}
			return false
		},
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		CreateFunc:  func(e event.CreateEvent) bool { return false },
		GenericFunc: func(ge event.GenericEvent) bool { return false },
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconcilerUpdates) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		WithEventFilter(ignoreEverythingButUpdatesPredicate(r.PodTargetContains)).
		Complete(r)
}
