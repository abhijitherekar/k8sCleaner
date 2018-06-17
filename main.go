package main

import (
	"fmt"
	"github.com/k8sCleaner/cmd"
)

func main() {
	fmt.Println("starting k8s-cleanner")
	cmd.Execute()
}

/*
package main

import (
	"fmt"
	"os"
	"github.com/chzyer/readline"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := os.Getenv("HOME") + "/.kube/config"
	fmt.Println("kubeconfig: ", kubeconfig)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	rl, err := readline.New("k8s> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil || line == "exit" {
			break
		}
		if line == "ps" {
			pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
			if err != nil {
				panic(err.Error())
			}
			for _, pod := range pods.Items {
				fmt.Printf("%s %s\n", pod.GetName(), pod.GetCreationTimestamp())
			}
		} else {
			fmt.Printf("unknown command\n")
		}
	}
}

*/
