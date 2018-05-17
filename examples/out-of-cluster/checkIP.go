/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/pkg/api/v1"
	"strings"
)
var	occupied_ip = []string{"10.30.99.52","10.30.99.62","172.25.12.128","172.25.12.129","172.25.12.131","172.25.12.134","172.25.12.136","172.25.12.138","172.25.12.139","172.25.12.140","172.25.12.142","172.25.12.144","172.25.12.150","172.25.12.151","172.25.12.154","172.25.12.162","172.25.12.163","172.25.12.164","172.25.12.166","172.25.12.167","172.25.12.174","172.25.12.135","172.25.12.147","172.25.12.179","172.25.12.183","172.25.12.188","172.25.12.197","172.25.12.176","172.25.12.175","172.25.12.137","172.25.12.190","172.25.12.199","172.25.12.192","172.25.12.156","172.25.12.145","172.25.12.211","172.25.12.212","172.25.12.213","172.25.12.214","172.25.12.217","172.25.12.219","172.25.12.228","172.25.12.230","172.25.12.234","172.25.12.239","172.25.12.242","172.25.12.244","172.25.12.251","172.25.12.253","172.25.12.254","10.30.99.56","172.25.12.157","172.25.12.180","172.25.12.194","172.25.12.222","172.25.12.223","172.25.12.225","172.25.12.226","172.25.12.186","10.30.99.54","172.25.12.146","172.25.12.172","172.25.12.204","172.25.12.189","172.25.12.205","172.25.12.224","172.25.12.181","172.25.12.229","172.25.12.232","172.25.12.240","172.25.12.209","172.25.12.231","172.25.12.241","172.25.12.216","172.25.12.243","172.25.12.210","172.25.12.198","172.25.12.215","172.25.12.182","172.25.12.252","172.25.12.245","172.25.12.132","172.25.12.155","172.25.12.160","172.25.12.206","172.25.12.246","172.25.13.128","172.25.13.139","172.25.13.140","172.25.13.141","172.25.13.142","172.25.13.143","172.25.13.144","172.25.13.145","172.25.13.148","172.25.12.250","172.25.13.150","172.25.13.167","172.25.13.160","172.25.13.161","172.25.13.168","172.25.12.187","172.25.12.218","172.25.13.175","172.25.13.183","172.25.13.184","172.25.13.185","172.25.13.187","172.25.13.188","172.25.13.189","172.25.13.190","172.25.13.191","172.25.13.192","172.25.13.194","172.25.12.193","172.25.13.200","172.25.13.210","172.25.13.211","172.25.13.212","172.25.12.149","172.25.13.169","172.25.13.171","172.25.13.172","172.25.13.173","172.25.13.174","172.25.13.176","172.25.13.177","172.25.13.178","172.25.13.181","172.25.13.182","172.25.13.195","172.25.13.197","172.25.13.198","172.25.13.217","172.25.13.202","172.25.13.214","172.25.13.215","172.25.13.216","172.25.13.219","172.25.12.221","172.25.13.166","172.25.13.220","172.25.12.236","172.25.13.179","172.25.13.180","172.25.13.199","172.25.13.213","172.25.12.143","172.25.13.223","172.25.12.153","172.25.13.162","172.25.13.163","172.25.13.164","172.25.13.170","172.25.13.186","172.25.12.152","172.25.13.193","172.25.13.208","172.25.13.218","172.25.13.165", "172.25.13.224","172.25.13.227","10.30.99.58","10.30.99.59","10.30.99.60","10.30.99.61","10.30.99.63","10.30.99.64","10.30.99.65","10.30.99.66","10.30.99.67","10.30.99.68"}

func containIP(pod *v1.Pod) bool {
	for _, ip := range occupied_ip {
		if strings.Contains(pod.Annotations["ips"], ip) {
			return true
		}
	}
	return false
}

func main() {
	fmt.Println("occupied_ip ", len(occupied_ip))

	kubeconfig := flag.String("kubeconfig", "./config", "absolute path to the kubeconfig file")
	flag.Parse()
	// uses the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	pods, err := clientset.CoreV1().Pods(os.Args[1]).List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, pod := range pods.Items {
		if pod.Annotations["ips"] == "" && pod.Labels["network"] != ""{
			fmt.Println(pod.Name)
		}
	}
	//for key, val := range occupied_map {
	//	if val == false {
	//		fmt.Println(pod.Name, key)
	//	}
	//}
}
