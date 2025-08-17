package display

import (
	"fmt"
	"strings"

	"github.com/start-codex/rke-update-checker/internal/rancher"
)

// PrintResults imprime los resultados en formato tabla
func PrintResults(apps []rancher.HelmApp) {
	if len(apps) == 0 {
		fmt.Println("No Helm applications found")
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 160))
	fmt.Printf("%-15s | %-12s | %-20s | %-15s | %-20s | %-12s | %-12s | %-8s | %-15s | %s\n",
		"CLUSTER", "NAMESPACE", "RELEASE", "REPO", "CHART", "CURRENT", "LATEST", "STATUS", "UPDATE", "SOURCES")
	fmt.Println(strings.Repeat("=", 160))

	updatesAvailable := 0

	for _, app := range apps {
		updateStatus := "✓ UP-TO-DATE"
		if app.LatestVersion == "managed" {
			updateStatus = "🔧 MANAGED"
		} else if app.LatestVersion == "internal" {
			updateStatus = "INTERNAL"
		} else if app.LatestVersion == "unknown" {
			updateStatus = "❓ NOT FOUND"
		} else if app.UpdateAvailable {
			updateStatus = "⚠ UPDATE AVAILABLE"
			updatesAvailable++
		}

		fmt.Printf("%-15s | %-12s | %-20s | %-15s | %-20s | %-12s | %-12s | %-8s | %-15s | %s\n",
			truncateString(app.Cluster, 15),
			truncateString(app.Release.Namespace, 12),
			truncateString(app.Release.Name, 20),
			truncateString(app.Release.ChartRepo, 15),
			truncateString(app.Release.ChartName, 20),
			truncateString(app.CurrentVersion, 12),
			truncateString(app.LatestVersion, 12),
			truncateString(app.Release.Status, 8),
			updateStatus,
			truncateString(strings.Join(app.Release.Sources, ", "), 50),
		)
	}

	fmt.Printf("\nTotal applications: %d\n", len(apps))
	fmt.Printf("Updates available: %d\n", updatesAvailable)
}

// truncateString trunca una string a una longitud máxima
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}