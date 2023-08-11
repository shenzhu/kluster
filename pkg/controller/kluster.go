package controller

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kanisterio/kanister/pkg/poll"
	"github.com/shenzhu/kluster/pkg/apis/shenzhu.dev/v1alpha1"
	klientset "github.com/shenzhu/kluster/pkg/client/clientset/versioned"
	kinf "github.com/shenzhu/kluster/pkg/client/informers/externalversions/shenzhu.dev/v1alpha1"
	klister "github.com/shenzhu/kluster/pkg/client/listers/shenzhu.dev/v1alpha1"
	"github.com/shenzhu/kluster/pkg/do"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	// clientset
	client kubernetes.Interface

	// clientset for custom resource kluster
	klient klientset.Interface

	// kluster has synced
	klusterSynced cache.InformerSynced

	// lister
	kLister klister.KlusterLister

	// queue
	wq workqueue.RateLimitingInterface

	recorder record.EventRecorder
}

func NewController(client kubernetes.Interface, klient klientset.Interface, klusterInformer kinf.KlusterInformer) *Controller {
	c := &Controller{
		client:        client,
		klient:        klient,
		klusterSynced: klusterInformer.Informer().HasSynced,
		kLister:       klusterInformer.Lister(),
		wq:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "kluster"),
	}

	klusterInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAdd,
			DeleteFunc: c.handleDelete,
		},
	)

	return c
}

func (c *Controller) Run(ch chan struct{}) error {
	if ok := cache.WaitForCacheSync(ch, c.klusterSynced); !ok {
		log.Println("cache was not synced")
	}

	go wait.Until(c.worker, time.Second, ch)

	<-ch
	return nil
}

func (c *Controller) worker() {
	for c.processNextItem() {

	}
}

func (c *Controller) processNextItem() bool {
	item, shutdDown := c.wq.Get()
	if shutdDown {
		// log as well
		return false
	}

	defer c.wq.Forget(item)

	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		log.Printf("error calling Namespace key func on cache for item: %v", err)
		return false
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Printf("splitting key into namespace and name, error: %v", err)
		return false
	}

	kluster, err := c.kLister.Klusters(ns).Get(name)
	if err != nil {
		log.Printf("error getting kluster from lister: %v", err)
		return false
	}

	log.Printf("processing kluster: %+v", kluster.Spec)

	clusterID, err := do.Create(c.client, kluster.Spec)
	if err != nil {
		log.Printf("error creating kluster: %v", err)
		return false
	}
	fmt.Printf("clusterID: %s\n", clusterID)

	err = c.updateStatus(clusterID, "creating", kluster)
	if err != nil {
		log.Printf("error updating status: %v", err)
		return false
	}

	err = c.waitForCluster(kluster.Spec, clusterID)
	if err != nil {
		log.Printf("error waiting for cluster: %v", err)
	}

	err = c.updateStatus(clusterID, "running", kluster)
	if err != nil {
		log.Printf("error updating status: %v", err)
	}

	return true
}

func (c *Controller) waitForCluster(spec v1alpha1.KlusterSpec, clusterID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	return poll.Wait(ctx, func(ctx context.Context) (bool, error) {
		state, err := do.ClusterState(c.client, spec, clusterID)
		if err != nil {
			return false, err
		}
		if state == "running" {
			return true, nil
		}

		return false, nil
	})
}

func (c *Controller) updateStatus(id string, progress string, kluster *v1alpha1.Kluster) error {
	// get the latest version of kluster
	k, err := c.klient.ShenzhuV1alpha1().Klusters(kluster.Namespace).Get(context.Background(), kluster.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	k.Status.KlusterID = id
	k.Status.Progress = progress
	_, err = c.klient.ShenzhuV1alpha1().Klusters(kluster.Namespace).UpdateStatus(context.Background(), k, metav1.UpdateOptions{})

	return err
}

func (c *Controller) handleAdd(obj interface{}) {
	log.Println("handleAdd was called")
	c.wq.Add(obj)
}

func (c *Controller) handleDelete(obj interface{}) {
	log.Println("handleDel was called")
	c.wq.Add(obj)
}
