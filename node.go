package main

import (
	"context"
	"encoding/json"
	"net/http"

	echo "github.com/labstack/echo/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	e := echo.New()

	// 创建一个in-cluster配置
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// 创建一个clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// 定义一个获取所有节点信息的接口
	e.GET("/api/nodes", func(c echo.Context) error {
		// 获取所有节点的列表
		nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		// 将节点列表转换为JSON格式并返回
		data, err := json.Marshal(nodes)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSONBlob(http.StatusOK, data)
	})

	// 定义一个获取所有命名空间信息的接口
	e.GET("/api/namespaces", func(c echo.Context) error {
		// 获取所有命名空间的列表
		namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		// 将命名空间列表转换为JSON格式并返回
		data, err := json.Marshal(namespaces)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSONBlob(http.StatusOK, data)
	})

	// 定义一个获取某个命名空间下所有pod信息的接口，需要传入命名空间名称作为参数
	e.GET("/api/pods/:namespace", func(c echo.Context) error {

		// 获取命名空间名称参数
		namespace := c.Param("namespace")

		// 获取该命名空间下所有pod的列表
		pods, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		// 将pod列表转换为JSON格式并返回
		data, err := json.Marshal(pods)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSONBlob(http.StatusOK, data)

	})

	e.Logger.Fatal(e.Start(":8080"))
}
