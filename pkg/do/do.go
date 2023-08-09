package do

import (
	"context"
	"fmt"
	"strings"

	"github.com/digitalocean/godo"
	"k8s.io/client-go/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/shenzhu/kluster/pkg/apis/shenzhu.dev/v1alpha1"
)

func Create(client kubernetes.Interface, spec v1alpha1.KlusterSpec) (string, error) {
	token, err := getToken(client, spec.TokenSecret)
	if err != nil {
		return "", err
	}

	dclient := godo.NewFromToken(token)
	fmt.Println(dclient)

	request := &godo.KubernetesClusterCreateRequest{
		Name:        spec.Name,
		RegionSlug:  spec.Region,
		VersionSlug: spec.Version,
		NodePools: []*godo.KubernetesNodePoolCreateRequest{
			{
				Size:  spec.NodePools[0].Size,
				Name:  spec.NodePools[0].Name,
				Count: spec.NodePools[0].Count,
			},
		},
	}

	fmt.Println(request)
	fmt.Println(request.NodePools[0])

	cluster, _, err := dclient.Kubernetes.Create(context.Background(), request)
	if err != nil {
		return "", err
	}

	return cluster.ID, nil
}

func getToken(client kubernetes.Interface, sec string) (string, error) {
	namespace := strings.Split(sec, "/")[0]
	name := strings.Split(sec, "/")[1]

	s, err := client.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return string(s.Data["token"]), nil
}
