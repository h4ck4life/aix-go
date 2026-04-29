package main

import (
	"fmt"
	"os"

	"github.com/h4ck4life/aix-go/cmd"
	"github.com/h4ck4life/aix-go/utils"
)

const version = "1.0.0"

func main() {
	if err := cmd.Execute(version); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(utils.ExitGeneralError)
	}
}
