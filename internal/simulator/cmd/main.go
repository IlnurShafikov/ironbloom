package main

import (
	"log"

	"github.com/IlnurShafikov/ironbloom/internal/simulator"
	"github.com/IlnurShafikov/ironbloom/internal/simulator/runners"
)

func main() {
	cfg := simulator.ParseConfigs()

	simManager := runners.NewSimulationManager()
	if err := simManager.RunSimulation(cfg); err != nil {
		//nolint:forbidigo // it's ok for simulator
		log.Fatal(err)
	}
}
