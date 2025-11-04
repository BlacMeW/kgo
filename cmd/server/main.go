package main

import (
	"flag"

	"k8s-dashboard/pkg/api"
	"k8s-dashboard/pkg/config"
	"k8s-dashboard/pkg/k8s"
	"k8s-dashboard/pkg/metrics"
	"k8s-dashboard/pkg/tui"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

func main() {
	configPath := flag.String("config", "", "path to configuration file")
	kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig file (overrides config file)")
	port := flag.String("port", "", "server port (overrides config file)")
	tuiMode := flag.Bool("tui", false, "run in terminal UI mode")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		klog.Fatalf("Failed to load config: %v", err)
	}

	// Override config with command line flags
	if *kubeconfig != "" {
		cfg.Kubernetes.Kubeconfig = *kubeconfig
	}
	if *port != "" {
		cfg.Server.Port = *port
	}

	clientset, err := k8s.NewClient(cfg.Kubernetes.Kubeconfig)
	if err != nil {
		klog.Fatalf("Failed to create k8s client: %v", err)
	}

	if *tuiMode {
		// Run TUI directly with clientset
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
		resourceHandler := api.NewResourceHandler(clientset)
		metricsHandler := metrics.NewMetricsHandler(clientset)

		r := gin.Default()
		r.Use(cors.Default())

		v1 := r.Group("/api/v1")
		{
			// Pod operations
			v1.GET("/pods", handler.ListPods)
			v1.POST("/pods/:namespace", handler.CreatePod)
			v1.PUT("/pods/:namespace/:name", handler.UpdatePod)
			v1.DELETE("/pods/:namespace/:name", handler.DeletePod)
			v1.GET("/pods/watch", handler.WatchPods)
			v1.GET("/pods/:namespace/:name/logs", resourceHandler.GetPodLogs)
			v1.GET("/pods/:namespace/:name/exec", resourceHandler.ExecPod)

			// Deployment operations
			v1.GET("/deployments", resourceHandler.ListDeployments)
			v1.POST("/deployments/:namespace", resourceHandler.CreateDeployment)
			v1.PUT("/deployments/:namespace/:name", resourceHandler.UpdateDeployment)
			v1.DELETE("/deployments/:namespace/:name", resourceHandler.DeleteDeployment)

			// Service operations
			v1.GET("/services", resourceHandler.ListServices)
			v1.POST("/services/:namespace", resourceHandler.CreateService)
			v1.PUT("/services/:namespace/:name", resourceHandler.UpdateService)
			v1.DELETE("/services/:namespace/:name", resourceHandler.DeleteService)

			// ConfigMap operations
			v1.GET("/configmaps", resourceHandler.ListConfigMaps)
			v1.POST("/configmaps/:namespace", resourceHandler.CreateConfigMap)
			v1.PUT("/configmaps/:namespace/:name", resourceHandler.UpdateConfigMap)
			v1.DELETE("/configmaps/:namespace/:name", resourceHandler.DeleteConfigMap)

			// Metrics operations
			v1.GET("/metrics/cluster", metricsHandler.GetClusterMetrics)
			v1.GET("/metrics/namespace/:namespace", metricsHandler.GetNamespaceMetrics)
		}

		klog.Info("Starting API server on :" + cfg.Server.Port)
		r.Run(":" + cfg.Server.Port)
	}
}
