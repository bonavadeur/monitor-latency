package main

import (
	"bytes"
	"context"
	"io"
	"strings"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	e := echo.New()
	e.GET("/metrics", func(c echo.Context) error {
		nodes, _ := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		podLog := make([]*rest.Request, len(nodes.Items))
		log := make([]string, len(nodes.Items))

		podName := make([]string, len(nodes.Items))
		for i := 1; i <= len(nodes.Items); i++ {
			podName[i-1] = "monitor-latency-agent-" + strconv.Itoa(i)
		}

		for i := 0; i < len(nodes.Items); i++ {
			podLog[i] = clientset.CoreV1().Pods("default").GetLogs(podName[i], &corev1.PodLogOptions{})
			podLogs, err := podLog[i].Stream(context.TODO())
			if err != nil {
				panic(err.Error())
			}
			defer podLogs.Close()

			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, podLogs)
			if err != nil {
				panic(err.Error())
			}
			str := buf.String()
			lines := strings.Split(str, "\n")
			log[i] = strings.Join(lines[len(lines)-len(nodes.Items):], "\n")
		}

		laslog := strings.Join(log, "")
		return c.String(http.StatusOK, laslog)
	})
	e.Logger.Fatal(e.Start(":9090"))
}
