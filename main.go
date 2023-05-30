package main

import (
	"context"
	"github.com/labstack/echo/v4"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
)

// Pod is a struct to receive the request body for creating a pod
type Pod struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Image     string `json:"image"`
}

// createInClusterConfig creates an in-cluster configuration for kubernetes client
func createInClusterConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return config, nil
}

// createClientset creates a clientset for kubernetes operations
func createClientset(config *rest.Config) (*kubernetes.Clientset, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

// listPods returns a handler function that gets all pods information and returns them as JSON
func listPods(clientset *kubernetes.Clientset) echo.HandlerFunc {
	return func(c echo.Context) error {
		pods, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		data, err := json.Marshal(pods)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.JSONBlob(http.StatusOK, data)
	}
}

// createPod returns a handler function that creates a new pod from the request body and returns it as JSON
func createPod(clientset *kubernetes.Clientset) echo.HandlerFunc {
	return func(c echo.Context) error {

		// Create a Pod instance and bind the request body data to it
		pod := new(Pod)
		if err := c.Bind(pod); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		// Create a kubernetes pod object from the request body data
		k8sPod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  pod.Name,
						Image: pod.Image,
					},
				},
			},
		}

		// Use Create method to create the pod
		result, err := clientset.CoreV1().Pods(pod.Namespace).Create(context.Background(), k8sPod, metav1.CreateOptions{})

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		// Return JSON format response
		data, _ := json.Marshal(result)

		return c.JSONBlob(http.StatusCreated, data)

	}
}

// createPod returns a handler function that creates a new pod from the request body and returns it as JSON
func updatePod(clientset *kubernetes.Clientset) echo.HandlerFunc {
	return func(c echo.Context) error {

		// Create a Pod instance and bind the request body data to it
		pod := new(Pod)
		if err := c.Bind(pod); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		// Create a kubernetes pod object from the request body data
		k8sPod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: pod.Name + "-", // Use GenerateName to avoid duplicate pod names
				Namespace:    pod.Namespace,
				Labels: map[string]string{ // Add some labels for easy identification and selection
					"app":  pod.Name,
					"role": "web",
				},
				Annotations: map[string]string{ // Add some annotations for extra information
					"createdBy": "Bing",
					"createdAt": time.Now().Format(time.RFC3339),
				},
			},
			//Spec: v1.PodSpec{
			//	RestartPolicy: v1.RestartPolicyAlways, // Set the restart policy to always restart the container on failure
			//	DNSPolicy:     v1.DNSClusterFirst,     // Set the DNS policy to use cluster DNS service first
			//	Containers: []v1.Container{
			//		{
			//			Name:  pod.Name,
			//			Image: pod.Image,
			//			LivenessProbe: &v1.Probe{ // Add a liveness probe to check if the container is alive
			//				FailureThreshold:    3,
			//				InitialDelaySeconds: 10,
			//				SuccessThreshold:    1,
			//				TimeOutSeconds:      5,
			//				HttpGetAction(&v1.HTTPGetAction{
			//					Path:   "/healthz",
			//					Port:   intstr.FromInt(8080),
			//					Scheme: v1.URISchemeHTTP,
			//				}),
			//			},
			//			ReadinessProbe: &v1.Probe{ // Add a readiness probe to check if the container is ready to serve requests
			//				FailureThreshold:    3,
			//				InitialDelaySeconds: 10,
			//				SuccessThreshold:    1,
			//				TimeOutSeconds:      5,
			//				HttpGetAction(&v1.HTTPGetAction{
			//					Path:   "/readyz",
			//					Port:   intstr.FromInt(8080),
			//					Scheme: v1.URISchemeHTTP,
			//				}),
			//			},
			//		},
			//	},
			//},
		}

		// Use Create method to create the pod
		result, err := clientset.CoreV1().Pods(pod.Namespace).Create(context.Background(), k8sPod, metav1.CreateOptions{})

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		// Return JSON format response without converting result to byte slice
		return c.JSON(http.StatusCreated, result)

	}
}

// getClusterStatus returns a handler function that gets the cluster status information and returns it as JSON
func getClusterStatus(clientset *kubernetes.Clientset) echo.HandlerFunc {
	return func(c echo.Context) error {

		// Get the cluster version information
		version, err := clientset.Discovery().ServerVersion()
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		// Get the cluster node count and health status
		nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		nodeCount := len(nodes.Items)
		nodeReadyCount := 0

		for _, node := range nodes.Items {
			for _, condition := range node.Status.Conditions {
				if condition.Type == "Ready" && condition.Status == "True" {
					nodeReadyCount++
				}
			}
		}

		// Get the cluster namespace count and name list
		namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		namespaceCount := len(namespaces.Items)

		namespaceNames := make([]string, 0)

		for _, namespace := range namespaces.Items {
			namespaceNames = append(namespaceNames, namespace.Name)

		}

		// Construct JSON format response data
		data := map[string]interface{}{
			"version":        version,
			"nodeCount":      nodeCount,
			"nodeReadyCount": nodeReadyCount,
			"namespaceCount": namespaceCount,
			"namespaceNames": namespaceNames,
		}

		// Return JSON format response
		return c.JSON(http.StatusOK, data)

	}
}

// listNodes returns a handler function that gets all nodes information and returns them as JSON
func listNodes(clientset *kubernetes.Clientset) echo.HandlerFunc {
	return func(c echo.Context) error {

		// Get all nodes list
		nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		// Convert nodes list to JSON format and return
		data, err := json.Marshal(nodes)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSONBlob(http.StatusOK, data)

	}
}

// listNamespaces returns a handler function that gets all namespaces information and returns them as JSON
func listNamespaces(clientset *kubernetes.Clientset) echo.HandlerFunc {
	return func(c echo.Context) error {

		// Get all namespaces list
		namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		// Convert namespaces list to JSON format and return
		data, err := json.Marshal(namespaces)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSONBlob(http.StatusOK, data)

	}
}

// listPodsByNamespace returns a handler function that gets all pods information in a given namespace and returns them as JSON
func listPodsByNamespace(clientset *kubernetes.Clientset) echo.HandlerFunc {
	return func(c echo.Context) error {

		// Get namespace name parameter
		namespace := c.Param("namespace")

		// Get all pods list in the given namespace
		pods, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		// Convert pods list to JSON format and return
		data, err := json.Marshal(pods)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSONBlob(http.StatusOK, data)

	}
}

func main() {

	e := echo.New()

	// Create an in-cluster configuration and a clientset for kubernetes operations
	config, err := createInClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := createClientset(config)
	if err != nil {
		panic(err.Error())
	}

	// Define an endpoint for getting cluster status using the handler function
	e.GET("/api/cluster/status", getClusterStatus(clientset))

	e.GET("/api/nodes", listNodes(clientset))
	e.GET("/api/namespaces", listNamespaces(clientset))
	e.GET("/api/pods/:namespace", listPodsByNamespace(clientset))

	// Define two endpoints for listing and creating pods using the handler functions
	e.GET("/api/pods", listPods(clientset))
	e.POST("/api/pods", createPod(clientset))

	e.Logger.Fatal(e.Start(":1323"))
}
