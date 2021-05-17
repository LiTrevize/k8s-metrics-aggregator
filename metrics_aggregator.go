package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type MetricsAggregator struct {
	IC                *InfluxdbClient
	ClientSet         *kubernetes.Clientset
	LastAggregateTime map[string]time.Time
}

func NewMetricsAggregator() *MetricsAggregator {
	ic := NewInfluxdbClient()
	clientset := NewClientSet()
	return &MetricsAggregator{
		IC:                ic,
		ClientSet:         clientset,
		LastAggregateTime: make(map[string]time.Time, 10)}
}

func NewClientSet() *kubernetes.Clientset {
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
	return clientset
}

func (ma *MetricsAggregator) LastAggregateTimeForNode(nodeName string) time.Time {
	if lastTime, ok := ma.LastAggregateTime[nodeName]; ok {
		return lastTime
	}
	ma.LastAggregateTime[nodeName] = ma.IC.queryLastTime("node", nodeName)
	return ma.LastAggregateTime[nodeName]
}

func (ma *MetricsAggregator) AggregateFromContainer(namespace, podname, containername string, sinceTime time.Time) (time.Time, int) {
	req := ma.ClientSet.CoreV1().Pods(namespace).GetLogs(podname, &corev1.PodLogOptions{Container: containername, SinceTime: &metav1.Time{Time: sinceTime}})
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		fmt.Println("error in opening stream: ", err)
	}
	defer podLogs.Close()
	lineReader := bufio.NewReader(podLogs)
	newTime := sinceTime
	count := 0
	for {
		line, err := lineReader.ReadSlice('\n')

		if len(line) == 0 {
			break
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("error read line: ", err)
		}
		// write to InfluxDB
		mlog := ma.IC.WriteMetricsFromLogBytes(line)
		if mlog != nil {
			newTime = mlog.Time
			count++
		}

	}
	return newTime, count

}

func (ma *MetricsAggregator) AggregateOnce() {
	podlist, err := ma.ClientSet.CoreV1().Pods("monitoring").List(context.TODO(), metav1.ListOptions{LabelSelector: "app=metrics-logger"})
	if err != nil {
		panic(err.Error())
	}
	for _, pod := range podlist.Items {
		lastTime := ma.LastAggregateTimeForNode(pod.Spec.NodeName)
		for _, container := range pod.Status.ContainerStatuses {
			if container.Ready {
				newTime, count := ma.AggregateFromContainer(pod.Namespace, pod.Name, container.Name, lastTime)
				ma.LastAggregateTime[pod.Spec.NodeName] = newTime
				fmt.Printf("Aggregate %d metrics from: Node %s, Pod %s, Container %s\n", count, pod.Spec.NodeName, pod.Name, container.Name)
			}
		}
	}
}

func (ma *MetricsAggregator) Aggregate(interval time.Duration) {
	ma.AggregateOnce()
	for range time.Tick(interval) {
		ma.AggregateOnce()
	}
}
