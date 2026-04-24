package main

import (
	"github.com/aluprince/Vector-Vanguard/scripts"
	"fmt"
	"log"
)

func main() {
	fmt.Println("Starting Scraper...")
	leads, err := scripts.ScrapeLeads("Shortlet Lagos")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d leads!\n", len(leads))
}
