package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// storageClassName represents the name of the StorageClass
var storageClassName string

// isDefaultStorageClassAnnotation represents a StorageClass annotation that marks a class as the default StorageClass
const isDefaultStorageClassAnnotation = "storageclass.kubernetes.io/is-default-class"

//  patchStringValue specifies a patch operation for a string.
type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

// isDefaultAnnotation returns a boolean if the annotation is set
func isDefaultAnnotation(obj metav1.ObjectMeta) bool {
	if obj.Annotations[isDefaultStorageClassAnnotation] == "true" {
		return true
	}
	return false
}

// updateDefaultAnnotation sets the annotation of the Default StorageClass to false
func patchDefaultAnnotation(storageClassName string, clientset *kubernetes.Clientset) {

	payload := []patchStringValue{{
		Op: "replace",
		// Do not change the JSON Pointer ~
		// Reference: RFC 6901 https://tools.ietf.org/html/rfc6901#section-3
		Path:  "/metadata/annotations/storageclass.kubernetes.io~1is-default-class",
		Value: "false",
	}}
	payloadBytes, _ := json.Marshal(payload)

	_, err := clientset.StorageV1().StorageClasses().Patch(storageClassName, types.JSONPatchType, payloadBytes)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("StorageClass %s was successfully patched", storageClassName)
}

// durFromEnv converts an environment variable to a duration of time
func durFromEnv(env string, def time.Duration) time.Duration {
	val := os.Getenv(env)
	if val == "" {
		return def
	}
	r := val[len(val)-1]
	if r >= '0' || r <= '9' {
		val = val + "s" // assume seconds
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		log.Fatalf("failed to parse %q: %s", env, err)
	}
	return d
}

func main() {

	// creates a signal channel to be notified of OS signals
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGKILL,
	)

	// goroutine to exit when invoked
	stop := func() {
		log.Println("Shutting down")
		os.Exit(0)
	}

	// sets storageClassName from environment variable - if empty, it defaults to gp2
	storageClassName = os.Getenv("STORAGE_CLASS_NAME")
	if storageClassName == "" {
		storageClassName = "gp2"
	}

	// sets checkInterval from environment variable - if empty, it defaults to 20s
	checkInterval := durFromEnv("CHECK_INTERVAL", 20*time.Second)

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	for {
		// captures system signals and invokes the stop function
		select {
		case <-signalCh:
			stop()
		default:
		}

		// list all storageclasses
		storageClasses, err := clientset.StorageV1().StorageClasses().List(metav1.ListOptions{})

		if err != nil {
			log.Fatal(err)
		}

		// loops through all StorageClasses
		// check if a StorageClasse is default and if the StorageClasse name is the same
		// if both is true, the StorageClasse will be patched
		for _, storageClass := range storageClasses.Items {

			if isDefaultAnnotation(storageClass.ObjectMeta) && storageClass.Name == storageClassName {
				log.Printf("StoragaClass %s is default annotated", storageClassName)
				patchDefaultAnnotation(storageClassName, clientset)
			} else if !isDefaultAnnotation(storageClass.ObjectMeta) && storageClass.Name == storageClassName {
				log.Printf("StorageClass %s is not default annotated", storageClassName)
			}
		}

		// captures system signals and invokes the stop function
		// wait for the checkInterval to count down
		select {
		case <-signalCh:
			stop()
		case <-time.After(checkInterval):
		}
	}
}
