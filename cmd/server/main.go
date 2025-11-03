package main

import (
	"flag"

	"k8s-dashboard/pkg/api"
	"k8s-dashboard/pkg/k8s"
	"k8s-dashboard/pkg/tui"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig file")
	tuiMode := flag.Bool("tui", false, "run in terminal UI mode")
	flag.Parse()

	clientset, err := k8s.NewClient(*kubeconfig)
	if err != nil {
		klog.Fatalf("Failed to create k8s client: %v", err)
	}

	if *tuiMode {
		// Run TUI
		tui, err := tui.NewTUI(clientset)
		if err != nil {
			klog.Fatalf("Failed to create TUI: %v", err)
		}

		if err := tui.Run(); err != nil {
			klog.Fatalf("TUI error: %v", err)
		}
	} else {
		// Run web server
		handler := api.NewHandler(clientset)

		r := gin.Default()
		r.Use(cors.Default())

		// Serve static web UI files
		r.Static("/web", "./web")
		r.StaticFile("/", "./web/index.html")

		v1 := r.Group("/api/v1")
		{
			v1.GET("/pods", handler.ListPods)
			v1.POST("/pods/:namespace", handler.CreatePod)
			v1.PUT("/pods/:namespace/:name", handler.UpdatePod)
			v1.DELETE("/pods/:namespace/:name", handler.DeletePod)
			v1.GET("/pods/watch", handler.WatchPods)
		}

		klog.Info("Starting web server on :8080")
		klog.Info("Web UI available at: http://localhost:8080")
		r.Run(":8080")
	}
}
