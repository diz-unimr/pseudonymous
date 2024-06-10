package main

import (
	"errors"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"log/slog"
	"os"
	"pseudonymous/config"
	"pseudonymous/fhir"
	"strings"
	"time"
)

func main() {

	var (
		projectName       string
		autoCreateDomains = true
	)

	appConfig, err := config.LoadConfig("./app.yaml")
	if err != nil {
		slog.Error("Error loading config file", "error", err.Error())
		os.Exit(1)
	}

	// create forms
	form := huh.NewForm(
		huh.NewGroup(
			// project name
			huh.NewInput().
				Title("What is the name of the project to create pseudonyms for?").
				Prompt("? ").
				Validate(notEmpty).
				Value(&projectName),

			// auto-create gPAS domains
			huh.NewConfirm().
				Title("Should we automatically create gPAS domains for pseudonyms of different data elements?").
				Value(&autoCreateDomains),
		),
	)

	err = form.Run()
	if err != nil {
		slog.Error("Failed to run form", "error", err.Error())
		os.Exit(1)
	}

	start := time.Now()
	err = spinner.New().
		Title("Running pseudonymization ...").
		Action(func() {
			p := fhir.NewProcessor(appConfig, projectName)
			err = p.Run()
			if err != nil {
				slog.Error("Processor run exited", "error", err.Error())
			}
		}).
		Run()

	if err != nil {
		slog.Error("Failed to run pseudonymization", "error", err.Error())
	} else {
		// TODO
		slog.Info("Pseudonymization finished", "time", time.Since(start))
	}

}

func notEmpty(s string) error {
	if strings.TrimSpace(s) == "" {
		return errors.New("input is empty")
	}

	return nil
}
