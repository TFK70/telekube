package controller

import (
	"context"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/client-go/informers"
  "k8s.io/klog"
)

const (
  Create = 0
  Update = 1
  Delete = 2
)

type EventType int32

type Event struct {
  Key string
  Type EventType
  Obj interface{}
  OldObj interface{}
  NewObj interface{}
}

type ControllerEventHandler func(clientSet kubernetes.Clientset, event Event) error

type Controller struct {
  clientSet kubernetes.Clientset
 
  Name string
  InformerFactory informers.SharedInformerFactory
  Queue workqueue.RateLimitingInterface
  EventHandlers []ControllerEventHandler
}

type ControllerOpts struct {
  Name string
  Namespace string
}

func New(opts ControllerOpts) Controller {
  config, err := rest.InClusterConfig()
  if err != nil {
    klog.Errorln("Could not retrieve config", err)
    os.Exit(1)
  }

  clientset, err := kubernetes.NewForConfig(config)
  if err != nil {
    klog.Errorln("Could not create client", err)
    os.Exit(1)
  }

  namespaceRestrictedInformerFactory := informers.WithNamespace(opts.Namespace)
  informerFactory := informers.NewSharedInformerFactoryWithOptions(clientset, 0, namespaceRestrictedInformerFactory)

  queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

  deploymentInformer :=  informerFactory.Apps().V1().Deployments()

  deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
    AddFunc: func(obj interface{}) {
      key, err := cache.MetaNamespaceKeyFunc(obj)
      if err == nil {
        queue.Add(Event{
          Key: key,
          Type: Create,
          Obj: obj,
        })
      }
    },
    UpdateFunc: func(oldObj, newObj interface{}) {
      key, err := cache.MetaNamespaceKeyFunc(newObj)
      if err == nil {
        queue.Add(Event{
          Key: key,
          Type: Update,
          OldObj: oldObj,
          NewObj: newObj,
        })
      }
    },
    DeleteFunc: func(obj interface{}) {
      key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
      if err == nil {
        queue.Add(Event{
          Key: key,
          Type: Delete,
          Obj: obj,
        })
      }
    },
  })
  
  return Controller{
    Name: opts.Name,
    InformerFactory: informerFactory,
    Queue: queue,
    EventHandlers: make([]ControllerEventHandler, 0),
  }
}

func (c *Controller) StartInformerFactory() {
  c.InformerFactory.Start(context.Background().Done())
}

func (c *Controller) AddEventHandler(eventHandler ControllerEventHandler) {
  c.EventHandlers = append(c.EventHandlers, eventHandler)
}

func (c *Controller) StartLoop() {
  klog.Infoln("Started controller loop")
  go func() {
    for {
      event, quit := c.Queue.Get()
      if quit {
        return
      }

      for _, eventHandler := range c.EventHandlers {
        err := eventHandler(c.clientSet, event.(Event))

        if err != nil {
          klog.Errorln("Handler ended up with an error", err)
        }
      }

      c.Queue.Forget(event)
      c.Queue.Done(event)
    }
  }()

  select{}
}
