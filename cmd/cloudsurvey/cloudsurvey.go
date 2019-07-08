package main

import (
	"context"
	"flag"
	"github.com/tetratom/cloudsurvey/pkg/config"
	"github.com/tetratom/cloudsurvey/pkg/core"
	"github.com/tetratom/cloudsurvey/pkg/metric"
	_ "github.com/tetratom/cloudsurvey/plugins"
	"golang.org/x/sync/errgroup"
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

var (
	newline = []byte{'\n'}
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

	ch := make(chan metric.Datum, 100)
	eg, c := errgroup.WithContext(context.Background())
	w := os.Stdout

	start := time.Now()

	eg.Go(func() error {
		defer close(ch)
		return runner.Run(c, ch)
	})

	eg.Go(func() error {
		for datum := range ch {
			select {
			case <-c.Done():
				return c.Err()
			default:
			}

			wire, err := datum.ToInfluxDBWireProtocol()
			if err != nil {
				return err
			}

			if _, err := w.Write([]byte(wire)); err != nil {
				return err
			}

			if _, err := w.Write(newline); err != nil {
				return err
			}
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Fatalf("error: %+v", err)
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
