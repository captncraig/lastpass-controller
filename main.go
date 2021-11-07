package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ansd/lastpass-go"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := os.Getenv("KUBECONFIG")
	var config *rest.Config
	var err error
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	factory := informers.NewSharedInformerFactoryWithOptions(clientset, 15*time.Minute,
		informers.WithTweakListOptions(func(lo *v1.ListOptions) {
			lo.LabelSelector = "lastpass-secret=true"
		}))
	informer := factory.Core().V1().ConfigMaps().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	fn := func(obj interface{}) {
		cm := obj.(*corev1.ConfigMap)
		log.Printf("New Lastpass template secret: %s", cm.GetName())
		sec := &corev1.Secret{}
		sec.SetName(cm.GetName())
		sec.Data = map[string][]byte{}
		sec.SetNamespace(cm.GetNamespace())
		var t = true
		sec.GetObjectMeta().SetOwnerReferences([]v1.OwnerReference{
			{
				Kind:       "ConfigMap",
				APIVersion: "v1",
				UID:        cm.GetUID(),
				Name:       cm.GetName(),
				Controller: &t,
			},
		})
		for k, v := range cm.Data {
			log.Println(k, v)
			sec.Data[k] = []byte(v)
		}
		log.Println(sec)
		log.Println(clientset.CoreV1().Secrets(cm.Namespace).Create(context.Background(), sec, v1.CreateOptions{}))

	}
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fn(obj)
		},
		UpdateFunc: func(old interface{}, obj interface{}) {
			fn(obj)
		},
	})
	informer.Run(stopper)
}

func getSecretData(name string) (string, error) {
	client, err := lastpass.NewClient(context.Background(), os.Getenv("LASTPASS_USER"), os.Getenv("LASTPASS_PASS"))
	if err != nil {
		return "", err
	}
	accounts, err := client.Accounts(context.Background())
	if err != nil {
		return "", err
	}
	for _, acc := range accounts {
		if acc.Name == name {
			return acc.Password, nil
		}
	}
	return "", fmt.Errorf("Account named %s not found", name)
}
