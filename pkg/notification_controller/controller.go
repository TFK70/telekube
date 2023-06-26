package notification_controller

import (
	"fmt"
	"net/http"
	"net/url"

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"telekube/internal/controller"
)

type NotificationControllerOpts struct {
	BotToken string
	ChatId   string
}

const (
	base_telegram_url = "https://api.telegram.org"
)

func NewNotificationController(opts NotificationControllerOpts) controller.Controller {
	nc := controller.New(controller.ControllerOpts{
		Name:      "notification-controller",
		Namespace: "default",
	})

	nc.StartInformerFactory()

	eventHandler := func(clientSet kubernetes.Clientset, event controller.Event) error {
		klog.Infoln(fmt.Sprintf("Handling %v event of %s", event.Type, event.Key))

		eventVerbMap := map[int32]string{
			0: "created",
			1: "updated",
			2: "deleted",
		}

		text := fmt.Sprintf("Deployment %s was %s", event.Key, eventVerbMap[int32(event.Type)])
		url := fmt.Sprintf("%s/bot%s/sendMessage?text=%s&chat_id=%s", base_telegram_url, opts.BotToken, url.QueryEscape(text), opts.ChatId)
		_, err := http.Get(url)
		if err != nil {
			return err
		}

		return nil
	}

	nc.AddEventHandler(eventHandler)

	return nc
}
