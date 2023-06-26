package main

import (
	"os"

	"telekube/pkg/notification_controller"
	"telekube/internal/telegram"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func main() {
  kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

  Namespace, _, err := kubeconfig.Namespace()
  if err != nil {
    klog.Errorln("Failed to get current namespace", err)
    os.Exit(1)
  }

  cfg, err := config.GetConfig()
  if err != nil {
    klog.Errorln("Failed to get kubernetes config", err)
    os.Exit(1)
  }

  mgr, err := manager.New(cfg, manager.Options{
    MetricsBindAddress: "0",
    LeaderElection: true,
    LeaderElectionID: "notification-controller",
    LeaderElectionNamespace: Namespace,
  })
  if err != nil {
    klog.Errorln("Failed to initialize new manager", err)
    os.Exit(1)
  }

  if err := notification_controller.AddToManager(mgr, telegram.BotOptions{Token: os.Getenv("BOT_TOKEN"), ChatId: os.Getenv("CHAT_ID")}); err != nil {
    klog.Errorln("Failed to setup controllers", err)
    os.Exit(1)
  }

  klog.Infoln("Starting the Cmd")

  if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
    klog.Errorln("Manager exited non-zero", err)
    os.Exit(1)
  }

  // nc := notification_controller.NewNotificationController(notification_controller.NotificationControllerOpts{
  //   BotToken: os.Getenv("BOT_TOKEN"),
  //   ChatId: os.Getenv("CHAT_ID"),
  // })
  // nc.StartLoop()
}
