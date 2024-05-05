package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/mtratsiuk/b3/b3"
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
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	flag.Parse()

	log.Info(fmt.Sprintf("got args: verbose=%v, help=%v, rootPath=%v", verbose, help, rootPath))

	if help {
		flag.PrintDefaults()
	}

	b3app, err := b3.NewApp(b3.Params{
		Log: log,
		Verbose: verbose,
		RootPath: rootPath,
	})

	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	fmt.Printf("%v", b3app)
}
