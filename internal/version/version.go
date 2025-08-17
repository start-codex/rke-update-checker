package version

import (
	"regexp"
	"strconv"
	"strings"
)

// IsNewer compara dos versiones y retorna true si newVersion es más nueva que currentVersion
func IsNewer(currentVersion, newVersion string) bool {
	if newVersion == "unknown" || newVersion == "internal" || newVersion == "managed" {
		return false
	}

	current := parseVersion(currentVersion)
	newer := parseVersion(newVersion)

	// Comparar major, minor, patch
	for i := 0; i < 3; i++ {
		if newer[i] > current[i] {
			return true
		} else if newer[i] < current[i] {
			return false
		}
	}

	// Las versiones son iguales
	return false
}

// parseVersion extrae los números de versión de una string
func parseVersion(version string) []int {
	// Remover prefijos comunes como v, V
	cleanVersion := strings.TrimPrefix(strings.TrimPrefix(version, "v"), "V")

	// Regex para extraer números separados por puntos
	re := regexp.MustCompile(`^(\d+)(?:\.(\d+))?(?:\.(\d+))?`)
	matches := re.FindStringSubmatch(cleanVersion)

	if matches == nil {
		return []int{0, 0, 0}
	}

	result := make([]int, 3)
	for i := 1; i < len(matches) && i <= 3; i++ {
		if matches[i] != "" {
			if num, err := strconv.Atoi(matches[i]); err == nil {
				result[i-1] = num
			}
		}
	}

	return result
}

// Compare compara dos versiones y retorna:
// -1 si a < b
//  0 si a == b
//  1 si a > b
func Compare(a, b string) int {
	versionA := parseVersion(a)
	versionB := parseVersion(b)

	for i := 0; i < 3; i++ {
		if versionA[i] < versionB[i] {
			return -1
		} else if versionA[i] > versionB[i] {
			return 1
		}
	}

	return 0
}