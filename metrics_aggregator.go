package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type MetricsAggregator struct {
	IC *InfluxdbClient
}

func NewMetricsAggregator() *MetricsAggregator {
	ic := NewInfluxdbClient()
	return &MetricsAggregator{IC: ic}
}

func (ma *MetricsAggregator) Test() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	podlist, err := clientset.CoreV1().Pods("monitoring").List(context.TODO(), metav1.ListOptions{LabelSelector: "app=metrics-logger"})
	if err != nil {
		panic(err.Error())
	}
	for _, pod := range podlist.Items {
		for _, container := range pod.Status.ContainerStatuses {
			if container.Ready {
				var seconds int64 = 120
				req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{Container: container.Name, SinceSeconds: &seconds, Follow: true})
				podLogs, err := req.Stream(context.TODO())
				if err != nil {
					fmt.Println("error in opening stream")
				}
				for {
					buf := make([]byte, 2000)
					numBytes, err := podLogs.Read(buf)
					if numBytes == 0 {
						break
					}
					if err == io.EOF {
						break
					}
					if err != nil {
						break
					}
					message := string(buf[:numBytes])
					fmt.Print(message)
				}
				podLogs.Close()

			}
		}
	}

}
