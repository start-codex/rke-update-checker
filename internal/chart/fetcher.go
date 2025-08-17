package chart

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	rancherClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

// Fetcher maneja la obtención de charts desde repositorios
type Fetcher struct {
	client  *rancherClient.Client
	verbose bool
}

// CatalogRepoResponse estructura para la respuesta de repositorios
type CatalogRepoResponse struct {
	Data []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Links struct {
			Index string `json:"index"`
		} `json:"links"`
	} `json:"data"`
}

// HelmIndexResponse estructura para la respuesta del índice de Helm
type HelmIndexResponse struct {
	Entries map[string][]struct {
		Home    string   `json:"home"`
		Name    string   `json:"name"`
		Version string   `json:"version"`
		Sources []string `json:"sources"`
	} `json:"entries"`
}

// NewFetcher crea un nuevo fetcher de charts
func NewFetcher(client *rancherClient.Client, verbose bool) *Fetcher {
	return &Fetcher{
		client:  client,
		verbose: verbose,
	}
}

// GetAllAvailableCharts obtiene todos los charts disponibles de todos los repositorios
func (f *Fetcher) GetAllAvailableCharts(cluster rancherClient.Cluster) ([]Chart, error) {
	repoData, err := f.getClusterReposWithLinks(cluster)
	if err != nil {
		return nil, fmt.Errorf("error getting cluster repos: %w", err)
	}

	var allCharts []Chart

	for _, repo := range repoData.Data {
		if repo.Links.Index != "" {
			charts, err := f.getChartsFromRepo(repo.ID, repo.Links.Index)
			if err != nil {
				continue
			}
			allCharts = append(allCharts, charts...)
		}
	}

	return allCharts, nil
}

// getChartsFromRepo obtiene todos los charts de un repositorio específico
func (f *Fetcher) getChartsFromRepo(repoID, indexURL string) ([]Chart, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", indexURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+f.client.Opts.TokenKey)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var indexResponse HelmIndexResponse
	if err := json.Unmarshal(body, &indexResponse); err != nil {
		return nil, err
	}

	var charts []Chart
	for chartName, versions := range indexResponse.Entries {
		if len(versions) > 0 {
			v := versions[0] // Primera versión = más reciente
			charts = append(charts, Chart{
				Version: v.Version,
				Repo:    repoID,
				Chart:   chartName,
				Name:    v.Name,
				Home:    v.Home,
				Sources: v.Sources,
			})
		}
	}

	return charts, nil
}

// getClusterReposWithLinks obtiene repositorios usando k8s.io/client-go
func (f *Fetcher) getClusterReposWithLinks(cluster rancherClient.Cluster) (CatalogRepoResponse, error) {
	kubeConfigAction, err := f.client.Cluster.ActionGenerateKubeconfig(&cluster)
	if err != nil {
		return CatalogRepoResponse{}, fmt.Errorf("error getting kubeconfig: %w", err)
	}

	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeConfigAction.Config))
	if err != nil {
		return CatalogRepoResponse{}, fmt.Errorf("error creating kube config: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return CatalogRepoResponse{}, fmt.Errorf("error creating dynamic client: %w", err)
	}

	clusterRepoGVR := schema.GroupVersionResource{
		Group:    "catalog.cattle.io",
		Version:  "v1",
		Resource: "clusterrepos",
	}

	listOptions := metav1.ListOptions{}
	repoList, err := dynamicClient.Resource(clusterRepoGVR).List(context.Background(), listOptions)
	if err != nil {
		return CatalogRepoResponse{}, fmt.Errorf("error listing clusterrepos: %w", err)
	}

	return f.convertUnstructuredToRepos(repoList), nil
}

// convertUnstructuredToRepos convierte la respuesta Unstructured a CatalogRepoResponse
func (f *Fetcher) convertUnstructuredToRepos(repoList *unstructured.UnstructuredList) CatalogRepoResponse {
	var response CatalogRepoResponse

	for _, item := range repoList.Items {
		name, found, err := unstructured.NestedString(item.Object, "metadata", "name")
		if !found || err != nil {
			continue
		}

		// Construir links basados en el patrón de Rancher
		catalogBaseURL := strings.Replace(f.client.Opts.URL, "/v3", "/v1/catalog.cattle.io.clusterrepos", 1)
		baseURL := catalogBaseURL + "/" + name

		repo := struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Links struct {
				Index string `json:"index"`
			} `json:"links"`
		}{
			ID:   name,
			Name: name,
		}

		repo.Links.Index = baseURL + "?link=index"
		response.Data = append(response.Data, repo)
	}

	return response
}