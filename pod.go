package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
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

	// 定义一个获取所有pod信息的GET接口
	e.GET("/api/pods", func(c echo.Context) error {
		// 获取所有pod的列表
		pods, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		// 将pod列表转换为JSON格式并返回
		data, err := json.Marshal(pods)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSONBlob(http.StatusOK,data)
	})

	// 定义一个创建pod的POST接口，需要传入JSON格式的请求体
	e.POST("/api/pods", func(c echo.Context) error {

		// 定义一个结构体来接收请求体中的数据
		type Pod struct {
			Name string `json:"name"`
			Namespace string `json:"namespace"`
			Image string `json:"image"`
		}

		// 创建一个Pod实例，并绑定请求体中的数据到该实例上
		pod := new(Pod)
		if err := c.Bind(pod); err != nil {
			return c.String(http.StatusBadRequest,err.Error())
		}

		// 根据请求体中的数据创建一个kubernetes pod对象
		k8sPod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: pod.Name,
				Namespace: pod.Namespace,
			},
			//Spec: v1.PodSpec{
			//	Containers: []v1.Container{
			//		{
			//			//Name: pod.Name,
			//			//Image: pod.Image,
			//		},
			//	},
			//},
		}

		// 使用Create方法创建pod
		result, err := clientset.CoreV1().Pods(pod.Namespace).Create(context.Background(), k8sPod, metav1.CreateOptions{})

		if err != nil {
			return c.String(http.StatusInternalServerError,err.Error())
		}

		// 返回JSON格式的响应
		data, _ := json.Marshal(result)

		return c.JSONBlob(http.StatusCreated,data)

	})

	// 定义一个更新pod信息的PUT接口，需要传入pod名称和命名空间作为参数和JSON格式的请求体
	e.PUT("/api/pods/:name/:namespace", func(c echo.Context) error {

		// 获取pod名称和命名空间参数
		name := c.Param("name")
		namespace := c.Param("namespace")

		// 定义一个结构体来接收请求体中的数据
		type Pod struct {
			Image string `json:"image"`
		}

		// 创建一个Pod实例，并绑定请求体中的数据到该实例上
		pod := new(Pod)
		if err := c.Bind(pod); err != nil {
			return c.String(http.StatusBadRequest,err.Error())
		}

		// 根据请求体中的数据创建一个kubernetes patch对象，用于更新指定字段
		patchData := fmt.Sprintf(`{"spec":{"containers":[{"name":"%s","image":"%s"}]}}`, name, pod.Image)

		patchBytes := []byte(patchData)

		patchType:= types.StrategicMergePatchType

		// 使用Patch方法更新pod
		result, err := clientset.CoreV1().Pods(namespace).Patch(context.Background(), name, patchType, patchBytes, metav1.PatchOptions{})

		if err != nil {
			return c.String(http.StatusInternalServerError,err.Error())
		}

		// 返回JSON格式的响应
		data, _ := json.Marshal(result)

		return c.JSONBlob(http.StatusOK,data)

	})

	// 定义一个删除pod信息的DELETE接口，需要传入pod名称和命名空间作为参数
	e.DELETE("/api/pods/:name/:namespace", func(c echo.Context) error {

		// 获取pod名称和命名空间参数
		name := c.Param("name")
		namespace := c.Param("namespace")

		// 使用Delete方法删除pod
		err := clientset.CoreV1().Pods(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})

		if err != nil {
			return c.String(http.StatusInternalServerError,err.Error())
		}

		// 返回空响应和204状态码表示删除成功
		return c.NoContent(http.StatusNoContent)

	})
