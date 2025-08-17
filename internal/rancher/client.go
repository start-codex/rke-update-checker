package rancher

import (
	"context"
	"fmt"
	"log"

	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/types"
	rancherClient "github.com/rancher/rancher/pkg/client/generated/management/v3"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/start-codex/rke-update-checker/internal/chart"
	"github.com/start-codex/rke-update-checker/internal/helm"
	"github.com/start-codex/rke-update-checker/internal/version"
)

// Config contiene la configuración para el cliente Rancher
type Config struct {
	URL     string
	Token   string
	Verbose bool
}

// Client encapsula el cliente de Rancher y funcionalidad relacionada
type Client struct {
	client  *rancherClient.Client
	config  *Config
}

// HelmApp representa una aplicación Helm con información de actualización
type HelmApp struct {
	Release         helm.Release
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	Cluster         string
}

// NewClient crea un nuevo cliente de Rancher
func NewClient(config *Config) (*Client, error) {
	client, err := rancherClient.NewClient(&clientbase.ClientOpts{
		URL:      config.URL,
		TokenKey: config.Token,
		Insecure: true,
	})
	if err != nil {
		return nil, fmt.Errorf("creating rancher client: %w", err)
	}

	return &Client{
		client: client,
		config: config,
	}, nil
}

// ListClusters lista todos los clusters disponibles
func (c *Client) ListClusters() ([]rancherClient.Cluster, error) {
	clusterList, err := c.client.Cluster.List(&types.ListOpts{})
	if err != nil {
		return nil, fmt.Errorf("listing clusters: %w", err)
	}
	return clusterList.Data, nil
}

// ProcessAllClusters procesa todos los clusters y retorna todas las aplicaciones Helm
func (c *Client) ProcessAllClusters(clusters []rancherClient.Cluster) ([]HelmApp, error) {
	var allApps []HelmApp

	for _, cluster := range clusters {
		if c.config.Verbose {
			log.Printf("Processing cluster: %s", cluster.Name)
		}

		apps, err := c.processCluster(cluster)
		if err != nil {
			log.Printf("Error processing cluster %s: %v", cluster.Name, err)
			continue
		}

		allApps = append(allApps, apps...)
	}

	return allApps, nil
}

// processCluster procesa un cluster individual
func (c *Client) processCluster(cluster rancherClient.Cluster) ([]HelmApp, error) {
	// Cargar charts disponibles una sola vez por cluster
	availableCharts, err := c.getAvailableCharts(cluster)
	if err != nil {
		if c.config.Verbose {
			log.Printf("Error loading available charts for cluster %s: %v", cluster.Name, err)
		}
		availableCharts = []chart.Chart{} // Fallback
	}

	// Obtener releases de Helm
	releases, err := c.getHelmReleases(cluster)
	if err != nil {
		return nil, fmt.Errorf("getting helm releases: %w", err)
	}

	// Procesar releases y calcular actualizaciones
	return c.processReleases(releases, availableCharts, cluster.Name), nil
}

// getAvailableCharts obtiene todos los charts disponibles del cluster
func (c *Client) getAvailableCharts(cluster rancherClient.Cluster) ([]chart.Chart, error) {
	chartFetcher := chart.NewFetcher(c.client, c.config.Verbose)
	return chartFetcher.GetAllAvailableCharts(cluster)
}

// getHelmReleases obtiene todos los releases de Helm del cluster
func (c *Client) getHelmReleases(cluster rancherClient.Cluster) ([]*helm.Release, error) {
	kubeConfigAction, err := c.client.Cluster.ActionGenerateKubeconfig(&cluster)
	if err != nil {
		return nil, fmt.Errorf("getting kubeconfig: %w", err)
	}

	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeConfigAction.Config))
	if err != nil {
		return nil, fmt.Errorf("creating kube config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating clientset: %w", err)
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing namespaces: %w", err)
	}

	var releases []*helm.Release
	seenReleases := make(map[string]*helm.Release)

	for _, ns := range namespaces.Items {
		secrets, err := clientset.CoreV1().Secrets(ns.Name).List(context.Background(), metav1.ListOptions{
			LabelSelector: "owner=helm",
		})
		if err != nil {
			continue
		}

		for _, secret := range secrets.Items {
			if releaseData, exists := secret.Data["release"]; exists {
				rel, err := helm.DecodeRelease(string(releaseData))
				if err != nil {
					continue
				}

				releaseKey := fmt.Sprintf("%s/%s", rel.Namespace, rel.Name)
				if existingRel, exists := seenReleases[releaseKey]; !exists || rel.Version > existingRel.Revision {
					helmRelease := helm.ExtractReleaseInfo(rel)
					seenReleases[releaseKey] = helmRelease
				}
			}
		}
	}

	for _, rel := range seenReleases {
		releases = append(releases, rel)
	}

	return releases, nil
}

// processReleases procesa releases y calcula información de actualizaciones
func (c *Client) processReleases(releases []*helm.Release, availableCharts []chart.Chart, clusterName string) []HelmApp {
	var apps []HelmApp

	for _, rel := range releases {
		latestVersion, repo := chart.FindLatestVersionBySource(rel.Sources, rel.ChartName, availableCharts)

		// Verificar si es chart interno/managed
		if isInternalChart(rel.ChartName) {
			latestVersion = "managed"
		}

		app := HelmApp{
			Release:         *rel,
			CurrentVersion:  rel.Version,
			LatestVersion:   latestVersion,
			UpdateAvailable: version.IsNewer(rel.Version, latestVersion),
			Cluster:         clusterName,
		}

		// Actualizar repo si se encontró
		if repo != "unknown" {
			app.Release.ChartRepo = repo
		}

		apps = append(apps, app)

		if c.config.Verbose {
			log.Printf("Chart=%s, Repo=%s, Current=%s, Latest=%s, Update=%v",
				rel.ChartName, app.Release.ChartRepo, rel.Version, latestVersion, app.UpdateAvailable)
		}
	}

	return apps
}