package updates_controller

import (
	"context"
	"os"
	"strings"
	"telekube/internal/telegram"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func AddToManager(mgr manager.Manager, botOpts telegram.BotOptions) error {
	bot, err := telegram.New(botOpts)
	if err != nil {
		klog.Errorln("Failed to create bot", err)
		os.Exit(1)
	}

	go bot.Start()

	return add(mgr, bot)
}

func getNamespacedName(key string) types.NamespacedName {
	sliced := strings.Split(key, "/")

	return types.NamespacedName{Name: sliced[1], Namespace: sliced[0]}
}

func add(mgr manager.Manager, bot telegram.Bot) error {
	client := mgr.GetClient()

	var deploymentHandler telegram.CommandHandler

	deploymentHandler = func(update telegram.Update) error {
		parts := strings.Split(update.Message.Text, " ")
		_, args := parts[0], parts[1:]

		klog.Infoln("Handling", args)

		if args[0] == "create" {
			dep := &appsv1.Deployment{}
			container := corev1.Container{}

			container.Name = args[1]
			container.Image = args[2]

			var replicas int32
			replicas = 1

			dep.ObjectMeta.Name = args[1]
			dep.Namespace = "default"
			dep.Spec.Selector = &metav1.LabelSelector{}
			dep.Spec.Selector.MatchLabels = make(map[string]string)
			dep.Spec.Selector.MatchLabels["app"] = args[1]
			dep.Spec.Replicas = &replicas
			dep.Spec.Template.ObjectMeta.Labels = make(map[string]string)
			dep.Spec.Template.ObjectMeta.Labels["app"] = args[1]
			dep.Spec.Template.Spec.Containers = append(dep.Spec.Template.Spec.Containers, container)

			if err := client.Create(context.TODO(), dep); err != nil {
				return err
			}
		} else if args[0] == "update" {
			dep := &appsv1.Deployment{}
			namespacedName := getNamespacedName(args[1])

			if err := client.Get(context.TODO(), namespacedName, dep); err != nil {
				return err
			}

			dep.Spec.Template.Spec.Containers[0].Name = args[2]
			dep.Spec.Template.Spec.Containers[0].Image = args[3]

			if err := client.Update(context.TODO(), dep); err != nil {
				return err
			}
		} else if args[0] == "delete" {
			dep := &appsv1.Deployment{}
			namespacedName := getNamespacedName(args[1])

			if err := client.Get(context.TODO(), namespacedName, dep); err != nil {
				return err
			}

			if err := client.Delete(context.TODO(), dep); err != nil {
				return err
			}
		}

		return nil
	}

	bot.AddHandler("/deployment", deploymentHandler)

	return nil
}
