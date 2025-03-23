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
var prod bool
var dry bool
var mode string

func init() {
	wd, err := os.Getwd()

	if err != nil {
		log.Panicf("failed to get working directory path: %v", err)
	}

	flag.StringVar(&rootPath, "root", wd, "path to the blog's root directory (folder containing 'b3.json')")
	flag.StringVar(&mode, "mode", "build", "'build' - build html files from markdown posts\n'cdn' - upload assets to cdn and replace urls in markdown files")
	flag.BoolVar(&verbose, "v", false, "verbose logging (debug)")
	flag.BoolVar(&help, "h", false, "print help (usage)")
	flag.BoolVar(&prod, "prod", false, "enable production build")
	flag.BoolVar(&dry, "dry", false, "execute in dry-run mode (preview affected assets before making actual CDN uploads)")
}

func main() {
	flag.Parse()

	logLevel := slog.LevelWarn

	if verbose {
		logLevel = slog.LevelDebug
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	log.Debug(
		fmt.Sprintf(`got args:
verbose=%v,
help=%v,
rootPath=%v,
prod=%v,
dry=%v,
mode=%v,
`,
			verbose,
			help,
			rootPath,
			prod,
			dry,
			mode,
		),
	)

	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	b3app, err := app.New(app.Params{
		Log:      log,
		Verbose:  verbose,
		RootPath: rootPath,
		Prod:     prod,
		DryRun:   dry,
	})

	if err != nil {
		log.Error(fmt.Sprintf("main: failed to create b3 app: %v", err))
		os.Exit(1)
	}

	var cmd func() error

	if mode == "cdn" {
		cmd = b3app.Cdn
	} else if mode == "build" {
		cmd = func() error {
			_, err := b3app.Build()
			return err
		}
	} else {
		log.Error(fmt.Sprintf("main: unexpected mode: %v", mode))
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := cmd(); err != nil {
		log.Error(fmt.Sprintf("main: failed to run b3: %v", err))
	}
}
