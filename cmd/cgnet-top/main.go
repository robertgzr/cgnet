package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/log/term"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/generic"

	"github.com/kinvolk/cgnet/pkg/service"
)

// "github.com/go-kit/kit/metrics/prometheus"
// stdprometheus "github.com/prometheus/client_golang/prometheus"

func run() int {
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s </path/to/cgroup>\n\n", os.Args[0])
		pflag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "")
	}

	verbose := pflag.BoolP("verbose", "v", false, "Enable verbose logging.")
	pflag.Parse()

	// Set up logging.
	logger := term.NewColorLogger(os.Stdout, log.NewLogfmtLogger, logColorFn)
	logger = log.With(logger, "ts", log.DefaultTimestamp)

	if *verbose {
		logger = level.NewFilter(logger, level.AllowDebug())
	} else {
		logger = level.NewFilter(logger, level.AllowInfo())
	}

	cgroupPath := pflag.Arg(0)
	if cgroupPath == "" {
		level.Error(logger).Log("msg", "No cgroup path given")
		return 1
	}

	var bytes, packets metrics.Counter
	{
		bytes = generic.NewCounter("bytes")
		packets = generic.NewCounter("packets")
	}

	ctx, cancel := context.WithCancel(context.Background())

	svc := service.New(logger)
	svc.AddCounterMetric(service.MetricPackets, packets)
	svc.AddCounterMetric(service.MetricBytes, bytes)

	if err := svc.Attach(ctx, cgroupPath); err != nil {
		level.Error(logger).Log("msg", "Attaching BPF program failed", "err", err)
		return 1
	}
	defer svc.Detach(ctx)

	term := make(chan os.Signal)
	signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-term:
			level.Info(logger).Log("msg", "Received SIGTERM, exiting gracefully...")
			cancel()

		case <-ctx.Done():
			level.Info(logger).Log("msg", "Exiting now")
			return 0

		case <-time.After(1 * time.Second):
			level.Info(logger).Log("msg", "New value", "bytes", bytes.(*generic.Counter).Value())
			level.Info(logger).Log("msg", "New value", "packets", packets.(*generic.Counter).Value())
		}
	}
}

func main() {
	os.Exit(run())
}

func logColorFn(keyvals ...interface{}) term.FgBgColor {
	for i := 1; i < len(keyvals)-1; i += 2 {
		if keyvals[i] != "level" {
			continue
		}
		switch keyvals[i+1] {
		case "error":
			return term.FgBgColor{Fg: term.Red}
		case "warn":
			return term.FgBgColor{Fg: term.Yellow}
		default:
			return term.FgBgColor{}
		}
	}
	return term.FgBgColor{}
}
