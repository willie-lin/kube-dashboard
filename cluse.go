package main

import (
	"context"
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

	// 定义一个获取集群状态的GET接口
	e.GET("/api/cluster/status", func(c echo.Context) error {
		// 获取集群版本信息
		version, err := clientset.Discovery().ServerVersion()
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		// 获取集群节点数量和健康状态
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

		// 获取集群命名空间数量和名称列表
		namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		namespaceCount := len(namespaces.Items)

		namespaceNames := make([]string, 0)

		for _, namespace := range namespaces.Items {
			namespaceNames = append(namespaceNames, namespace.Name)

		}

		// 构造JSON格式的响应数据
		data := map[string]interface{}{
			"version":        version,
			"nodeCount":      nodeCount,
			"nodeReadyCount": nodeReadyCount,
			"namespaceCount": namespaceCount,
			"namespaceNames": namespaceNames,
		}

		// 返回JSON格式的响应
		return c.JSON(http.StatusOK, data)

	})

	e.Logger.Fatal(e.Start(":8080"))
}