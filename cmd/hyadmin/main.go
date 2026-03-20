package main

import (
	"log"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "hyadmin",
		Short: "HySP Admin Platform Console",
	}
	root.AddCommand(serveCmd())
	root.AddCommand(migrateCmd())
	root.AddCommand(seedCmd())
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
