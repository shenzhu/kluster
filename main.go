package main

import (
	"flag"
	"log"
	"path/filepath"
	"time"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	klient "github.com/shenzhu/kluster/pkg/client/clientset/versioned"
	kInfoFac "github.com/shenzhu/kluster/pkg/client/informers/externalversions"
	"github.com/shenzhu/kluster/pkg/controller"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Printf("error building kubeconfig: %v", err)
	}

	klientset, err := klient.NewForConfig(config)
	if err != nil {
		log.Printf("error building klientset: %v", err)
	}

	infoFactory := kInfoFac.NewSharedInformerFactory(klientset, 20*time.Minute)

	ch := make(chan struct{})
	c := controller.NewController(klientset, infoFactory.Shenzhu().V1alpha1().Klusters())

	infoFactory.Start(ch)
	if err := c.Run(ch); err != nil {
		log.Printf("error running controller: %v", err.Error())
	}
}
