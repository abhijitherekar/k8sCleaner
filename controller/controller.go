package controller

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	api_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var ConfigFileName = "k8sCleaner-config.yaml"

// Resource contains resource configuration
type Resource struct {
	Deployment            bool `json:"deployment"`
	ReplicationController bool `json:"rc"`
	ReplicaSet            bool `json:"rs"`
	DaemonSet             bool `json:"ds"`
	Services              bool `json:"svc"`
	Pod                   bool `yaml:"pod"`
	Job                   bool `json:"job"`
	PersistentVolume      bool `json:"pv"`
}

// Config struct contains kubewatch configuration
type Config struct {
	//Reason   []string `json:"reason"`
	Resource Resource `json:"resource"`
}

// Load loads configuration from config file
func (c *Config) Load() error {

	file, err := os.Open(os.Getenv("HOME") + "/work/go/src/github.com/k8sCleaner/" + ConfigFileName)

	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	if len(b) != 0 {
		return yaml.Unmarshal(b, c)
	}

	return nil
}

// New creates new config object
func New() (*Config, error) {
	c := &Config{}
	if err := c.Load(); err != nil {
		return c, err
	}

	return c, nil
}

/*	for the controller structure we need the following:

	1. A Queue to process the updates on k8s resources
	2. A k8s-clientset to access all the resources of the cluster
	3. A shared index informer to listen on the resources
	4. A logger to log the activities of the controller
*/

type Controller struct {
	Clientset   kubernetes.Interface
	PodQueue    workqueue.RateLimitingInterface
	PodInformer cache.SharedIndexInformer
	//PodIndexer  cache.Indexer
	K8sConfig  *Config
	KubeConfig *rest.Config
}

func Start(clientset *kubernetes.Clientset, kubeconfig *rest.Config) {
	//load the config which tells which config to watch for
	k8sconfig, err := New()
	if err != nil {
		fmt.Println("\n error reading the config file")
	}

	//if pod is true start a PodInformer
	if k8sconfig.Resource.Pod {
		c := NewPodController(k8sconfig, clientset, kubeconfig)
		stopch := make(chan struct{})

		go c.Run(stopch)
		fmt.Println("creating sigtem signal")
		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGTERM)
		signal.Notify(sigterm, syscall.SIGINT)
		<-sigterm
		fmt.Println("\n\n Received the sigterm, close stopch")
		close(stopch)
	}
}

func NewPodController(k8sconfig *Config, client *kubernetes.Clientset, kubeconfig *rest.Config) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Pods(meta_v1.NamespaceAll).List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().Pods(meta_v1.NamespaceAll).Watch(options)
			},
		},
		&api_v1.Pod{},
		0, //Skip resync
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	})
	c := &Controller{
		Clientset:   client,
		PodQueue:    queue,
		PodInformer: informer,
		//		PodIndexer:  indexer,
		KubeConfig: kubeconfig,
		K8sConfig:  k8sconfig,
	}

	return c
}

// Run starts the k8sCleaner controller
func (c *Controller) Run(stopCh <-chan struct{}) {
	//defer utilruntime.HandleCrash()
	//defer c.PodQueue.ShutDown()

	fmt.Println("Starting  controller")

	go c.PodInformer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.PodInformer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	fmt.Println("Controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
	fmt.Println("Shutting down the controller")
}

func (c *Controller) runWorker() {
	for c.ProcessItem() {
	}
}

func (c *Controller) ProcessItem() bool {
	key, quit := c.PodQueue.Get()
	if quit {
		return false
	}
	defer c.PodQueue.Done(key)
	err := c.processPod(key.(string))
	if err == nil {
		c.PodQueue.Forget(key)
	} else {
		c.PodQueue.AddRateLimited(key)
	}
	return true
}

func (c *Controller) processPod(key string) error {
	podobj, present, err := c.PodInformer.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}
	if !present {
		fmt.Println("Pod deleted with key: ", key)
		return nil
	}
	fmt.Println("Pod added:", podobj)
	return nil
}
