package chart

// Chart representa un chart disponible en los repositorios
type Chart struct {
	Version string   `json:"version"`
	Repo    string   `json:"repo"`
	Chart   string   `json:"chart"`
	Name    string   `json:"name"`
	Home    string   `json:"home"`
	Sources []string `json:"sources"`
}

// FindLatestVersionBySource busca la versión más reciente de un chart por matching de Sources
func FindLatestVersionBySource(installedSources []string, chartName string, availableCharts []Chart) (string, string) {
	chart, found := FindChartByName(availableCharts, chartName, installedSources)
	if found {
		return chart.Version, chart.Repo
	}

	chart, found = FindChartBySource(availableCharts, installedSources)
	if found {
		return chart.Version, chart.Repo
	}

	return "unknown", "unknown"
}

// FindChartByName busca un chart por nombre con estrategia de fallback
func FindChartByName(charts []Chart, name string, sources []string) (Chart, bool) {
	// Try exact match first (name + sources)
	for _, chart := range charts {
		if (chart.Chart == name || chart.Name == name) && sourcesMatch(chart.Sources, sources) {
			return chart, true
		}
	}

	// Fallback to name-only match
	for _, chart := range charts {
		if chart.Chart == name || chart.Name == name {
			return chart, true
		}
	}

	return Chart{}, false
}

// FindChartBySource busca un chart únicamente por matching de sources
func FindChartBySource(charts []Chart, sources []string) (Chart, bool) {
	for _, chart := range charts {
		if sourcesMatch(chart.Sources, sources) {
			return chart, true
		}
	}
	return Chart{}, false
}

// sourcesMatch verifica si hay intersección entre dos slices de sources
func sourcesMatch(a, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}

	sourceSet := make(map[string]bool, len(a))
	for _, s := range a {
		sourceSet[s] = true
	}

	for _, s := range b {
		if sourceSet[s] {
			return true
		}
	}

	return false
}