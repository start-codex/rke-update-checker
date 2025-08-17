package rancher

import "strings"

// Charts internos que no están en repos públicos
var internalCharts = map[string]bool{
	"rancher-webhook":                true,
	"fleet-agent-local":              true,
	"rancher-provisioning-capi":      true,
	"system-upgrade-controller":      true,
	"rke2-canal":                     true,
	"rke2-coredns":                   true,
	"rke2-ingress-nginx":             true,
	"rke2-metrics-server":            true,
	"rke2-runtimeclasses":            true,
	"rke2-snapshot-controller":       true,
	"rke2-snapshot-controller-crd":   true,
}

// Prefijos de charts internos
var internalPrefixes = []string{
	"rke2-",
	"rancher-",
	"fleet-",
	"cattle-",
}

// isInternalChart verifica si un chart es administrado internamente
func isInternalChart(chartName string) bool {
	if internalCharts[chartName] {
		return true
	}

	for _, prefix := range internalPrefixes {
		if strings.HasPrefix(chartName, prefix) {
			return true
		}
	}

	return false
}