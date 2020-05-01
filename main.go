package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type handler struct{}

func (h *handler) OnAdd(obj interface{}) {
	log.Print("OnAdd")
}

func (h *handler) OnUpdate(oldObj, newObj interface{}) {
	log.Print("OnUpdate")
}
func (h *handler) OnDelete(obj interface{}) {
	log.Print("OnDelete")
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	stop := make(chan struct{})
	go func() {
		<-sigs
		close(stop)
	}()

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		log.Fatal("KUBECONFIG not defined")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("error reading kubeconfig: %v", err)
	}

	cliset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("error getting clientset: %v", err)
	}

	factory := informers.NewSharedInformerFactory(cliset, time.Minute)
	podsInformer := factory.Core().V1().Pods().Informer()
	podsInformer.AddEventHandlerWithResyncPeriod(&handler{}, time.Minute)

	go podsInformer.Run(stop)

	if !cache.WaitForCacheSync(stop, podsInformer.HasSynced) {
		log.Fatal("timed out waiting for caches to sync")
	}

	<-stop
}
