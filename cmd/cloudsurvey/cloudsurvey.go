package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"runtime/debug"
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
}

func version() string {
	info, ok := debug.ReadBuildInfo()
	if ok {
		version := info.Main.Version
		if version != "" {
			return info.Main.Version
		}
	}
	return "(devel)"
}
