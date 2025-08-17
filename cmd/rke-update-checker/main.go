package main

import (
	"log"
	"os"

	"github.com/start-codex/rke-update-checker/internal/display"
	"github.com/start-codex/rke-update-checker/internal/rancher"
)

func main() {
	// Configuración desde variables de entorno
	rancherURL := os.Getenv("RANCHER_URL")
	if rancherURL == "" {
		log.Fatal("RANCHER_URL environment variable is required")
	}

	token := os.Getenv("RANCHER_TOKEN")
	if token == "" {
		log.Fatal("RANCHER_TOKEN environment variable is required")
	}

	verbose := os.Getenv("VERBOSE") == "true"

	// Crear configuración
	config := &rancher.Config{
		URL:     rancherURL,
		Token:   token,
		Verbose: verbose,
	}

	// Crear cliente de Rancher
	client, err := rancher.NewClient(config)
	if err != nil {
		log.Fatalf("Error creating Rancher client: %v", err)
	}

	// Obtener lista de clusters
	clusters, err := client.ListClusters()
	if err != nil {
		log.Fatalf("Error listing clusters: %v", err)
	}

	if verbose {
		log.Printf("Found %d clusters", len(clusters))
	}

	// Procesar todos los clusters
	apps, err := client.ProcessAllClusters(clusters)
	if err != nil {
		log.Fatalf("Error processing clusters: %v", err)
	}

	// Mostrar resultados
	display.PrintResults(apps)
}