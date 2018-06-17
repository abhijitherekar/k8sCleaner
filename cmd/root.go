package cmd

import (
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "k8swatcher",
	Short: "watches the pods getting created",
	Long:  `A k8s watcher to watch for the pods getting created`,
	Run:   run_k8swatcher,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

/*	1. Now here at the run_k8swatcher, we need to create a kubernetes client
	this is got from the ~/.kube/config

	2. We need start the new custom controller which will listen to the
	pods
*/

func run_k8swatcher(cmd *cobra.Command, args []string) {
	kubeConfigPath = os.Getenv("HOME") + "/.kube/config"

	if _, err := os.Stat(kubeConfigPath); err == nil {
		config, err := clientcmd.BuildConfigFromFlags("", *kubeConfigPath)
		if err != nil {
			panic(err.Error())
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	for {
		pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	}

}