package notification_controller

import (
	"context"
	"fmt"
	"os"
	"telekube/internal/telegram"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func AddToManager(mgr manager.Manager, botOpts telegram.BotOptions) error {
	return add(mgr, newReconciler(mgr, botOpts))
}

func newReconciler(mgr manager.Manager, botOpts telegram.BotOptions) reconcile.Reconciler {
	bot, err := telegram.New(botOpts)
	if err != nil {
		klog.Errorln("Failed to create bot", err)
		os.Exit(1)
	}

	return &ReconcileJob{client: mgr.GetClient(), scheme: mgr.GetScheme(), bot: bot}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("notification-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	dep := &appsv1.Deployment{}
	err = c.Watch(source.Kind(mgr.GetCache(), dep), &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileJob{}

type ReconcileJob struct {
	client client.Client
	scheme *runtime.Scheme
	bot    telegram.Bot
}

func (r *ReconcileJob) Reconcile(context context.Context, request reconcile.Request) (reconcile.Result, error) {
	instance := &appsv1.Deployment{}
	err := r.client.Get(context, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		return reconcile.Result{}, err
	}

	message := fmt.Sprintf("Deployment changed: %s/%s", instance.Namespace, instance.Name)

	klog.Infoln(message)
	r.bot.Send(message)

	return reconcile.Result{}, nil
}
