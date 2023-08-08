package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	klient "github.com/shenzhu/kluster/pkg/client/clientset/versioned"
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

	klusters, err := klientset.ShenzhuV1alpha1().Klusters("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Printf("error listing klusters: %v", err)
	}
	fmt.Printf("length of klusters: %d\n", len(klusters.Items))
}
