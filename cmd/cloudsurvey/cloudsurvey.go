package main

import (
	"context"
	"flag"
	"github.com/tetratom/cloudsurvey/pkg/config"
	"github.com/tetratom/cloudsurvey/pkg/core"
	_ "github.com/tetratom/cloudsurvey/plugins"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

const (
	defaultConfigPath = "/etc/cloudsurvey/cloudsurvey.conf"
)

func main() {
	var opts struct {
		config  string
		verbose bool
	}

	flag.StringVar(&opts.config, "config", defaultConfigPath, "path to configuration file")
	flag.BoolVar(&opts.verbose, "verbose", false, "enable verbose output to stderr")
	flag.Parse()

	if opts.verbose {
		// it's default but let's be certain
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	log.Printf("cloudsurvey: %s %s", version(), runtime.Version())
	log.Printf("config: %s", opts.config)

	conf, err := config.FromFile(opts.config)
	if err != nil {
		log.Fatal("error: load config:", err)
	}

	runner, err := core.NewRunner(context.Background(), conf)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	if err := runner.Run(context.Background(), os.Stdout); err != nil {
		log.Fatal(err)
	}
	end := time.Now()
	log.Printf("elapsed %d ms", end.Sub(start).Nanoseconds()/1000000)
}

func version() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if version := info.Main.Version; version != "" {
			return version
		}
	}

	return "(devel)"
}
