package controllers

import (
	"context"
	"fmt"
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

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Slack             SlackConnector
	Scheme            *runtime.Scheme
	PodTargetContains string
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=get

// Reconcile reconciles changes to pod resources when they are created or deleted
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx).WithValues("pod", req.NamespacedName)
	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		if apierrors.IsNotFound(err) {
			l.Info("slackops", "DELETED", pod.Name)
			err = r.Slack.PostMessage(fmt.Sprintf("goodbye from %s", req.NamespacedName.Name))
			if err != nil {
				return ctrl.Result{}, err
			}
			// we'll ignore not-found errors, since we can get them on deleted requests.
			return ctrl.Result{}, nil
		}
		l.Error(err, "slackops unable to fetch Pod")
		return ctrl.Result{}, err
	}
	l.Info("slackops", "CREATED", pod.Name)
	err := r.Slack.PostMessage(fmt.Sprintf("hello from %s", pod.Name))
	return ctrl.Result{}, err
}

// predicates will filter if the functions evaluate to false
func ignoreNonTargetPodsPredicate(podTargetContains string) predicate.Predicate {
	return predicate.Funcs{
		DeleteFunc: func(e event.DeleteEvent) bool {
			return strings.Contains(e.Object.GetName(), podTargetContains)
		},
		CreateFunc: func(e event.CreateEvent) bool {
			// Only consider updates to pods containing our target pod name
			return strings.Contains(e.Object.GetName(), podTargetContains)
		},
		UpdateFunc:  func(e event.UpdateEvent) bool { return false },
		GenericFunc: func(ge event.GenericEvent) bool { return false },
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		WithEventFilter(ignoreNonTargetPodsPredicate(r.PodTargetContains)).
		Complete(r)
}
