package helm

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Release representa un release de Helm con información adicional
type Release struct {
	Name         string
	Namespace    string
	ChartName    string
	ChartRepo    string
	Version      string
	Status       string
	Revision     int
	Sources      []string
}

// DecodeRelease decodifica un secret de Helm en un release
func DecodeRelease(encoded string) (*release.Release, error) {
	// Paso 1: Base64 decode
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}

	// Paso 2: Gzip decompress
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("gzip decompress: %w", err)
	}
	defer gz.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, gz); err != nil {
		return nil, fmt.Errorf("copying data: %w", err)
	}

	// Paso 3: YAML/JSON decode
	var rel release.Release
	decoder := yaml.NewYAMLOrJSONDecoder(&buf, 4096)
	if err := decoder.Decode(&rel); err != nil {
		return nil, fmt.Errorf("yaml decode: %w", err)
	}

	return &rel, nil
}

// ExtractReleaseInfo extrae información relevante de un release de Helm
func ExtractReleaseInfo(rel *release.Release) *Release {
	chartRepo := "unknown"
	chartName := rel.Chart.Metadata.Name

	// Extraer repo del chart instalado si está en formato repo/chart
	if rel.Chart.Metadata.Name != "" {
		parts := strings.Split(rel.Chart.Metadata.Name, "/")
		if len(parts) == 2 {
			chartRepo = parts[0]
			chartName = parts[1]
		} else {
			// Intentar obtener repo de Sources si está disponible
			if len(rel.Chart.Metadata.Sources) > 0 {
				chartRepo = extractRepoFromSources(rel.Chart.Metadata.Sources)
			}
		}
	}

	return &Release{
		Name:      rel.Name,
		Namespace: rel.Namespace,
		ChartName: chartName,
		ChartRepo: chartRepo,
		Version:   rel.Chart.Metadata.Version,
		Status:    string(rel.Info.Status),
		Revision:  rel.Version,
		Sources:   rel.Chart.Metadata.Sources,
	}
}

// extractRepoFromSources intenta extraer el nombre del repo desde los sources
func extractRepoFromSources(sources []string) string {
	for _, source := range sources {
		if strings.Contains(source, "github.com") {
			// Extraer repo de URL como github.com/bitnami/charts
			urlParts := strings.Split(source, "/")
			if len(urlParts) >= 4 {
				return urlParts[3] // bitnami
			}
		}
	}
	return "unknown"
}