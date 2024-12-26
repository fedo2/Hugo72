package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	err := os.Chdir("phase2")
	if err != nil {
		log.Fatalf("Failed to change directory: %v", err)
	}

	cmd := exec.Command("hugo")
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Hugo build failed: %v", err)
	}
	log.Println("Hugo build succeeded")
}
