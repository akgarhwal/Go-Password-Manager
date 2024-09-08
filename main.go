package main

import (
	"context"
	"github.com/akgarhwal/go-password-manager/core"
	"github.com/akgarhwal/go-password-manager/util"
	"github.com/pterm/pterm"
)

func main() {

	pmContext := context.Background()

	util.ShowWelcomeText()

	pmContext, pm, err := core.LoadSavedPassword(pmContext)
	if err != nil {
		pterm.Error.Println("Error while starting Go Password Manager. Error: ", err)
		pterm.Error.Println("Exiting... :( ")
		return
	}

	core.StartInteractiveMode(pmContext, pm)
}
