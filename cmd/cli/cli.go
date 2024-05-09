package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/mtratsiuk/b3/pkg/app"
)

var verbose bool
var help bool
var rootPath string

func init() {
	wd, err := os.Getwd()

	if err != nil {
		log.Panicf("failed to get working directory path: %v", err)
	}

	flag.StringVar(&rootPath, "root", wd, "path to the blog's root directory (folder containing 'b3.json')")
	flag.BoolVar(&verbose, "v", false, "verbose logging (debug)")
	flag.BoolVar(&help, "h", false, "print help (usage)")
}

func main() {
	flag.Parse()

	logLevel := slog.LevelWarn

	if verbose {
		logLevel = slog.LevelDebug
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	log.Debug(fmt.Sprintf("got args: verbose=%v, help=%v, rootPath=%v", verbose, help, rootPath))

	if help {
		flag.PrintDefaults()
	}

	b3app, err := app.New(app.Params{
		Log:      log,
		Verbose:  verbose,
		RootPath: rootPath,
	})

	if err != nil {
		log.Error(fmt.Sprintf("main: failed to create b3 app: %v", err))
		os.Exit(1)
	}

	if err = b3app.Build(); err != nil {
		log.Error(fmt.Sprintf("main: failed to build: %v", err))
	}
}
