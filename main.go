package main

import (
	"fmt"
	"os"

	"github.com/bmeg/lathe/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
}
