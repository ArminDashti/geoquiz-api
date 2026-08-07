package main

import (
	"fmt"
	"os"

	"github.com/armin/geoquiz-api/internal/data"
)

func main() {
	store, err := data.LoadCountries("data/countries.geojson")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, c := range store.List() {
		if c.Name == "Germany" || c.Name == "France" || c.Name == "Japan" || c.Name == "Australia" {
			n, ok := store.Neighbors(c.ID)
			fmt.Printf("%s (%d) ok=%v neighbors=%d\n", c.Name, c.ID, ok, len(n))
			for _, x := range n {
				fmt.Printf("  - %s\n", x.Name)
			}
		}
	}
}
