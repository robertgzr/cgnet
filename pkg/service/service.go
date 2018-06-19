package service

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/metrics"
	"github.com/iovisor/gobpf/elf"
	"github.com/pkg/errors"

	"github.com/kinvolk/cgnet/pkg/bpf"
)

const (
	lookupPeriod = 1 * time.Second

	MetricPackets uint32 = 0
	MetricBytes   uint32 = 1
)

type CgnetService interface {
	Attach(ctx context.Context, cgroupPath string) error
	Detach(context.Context)
	AddCounterMetric(bpfkey uint32, counter metrics.Counter)
}

func New(logger log.Logger) CgnetService {
	return &cgService{
		logger:  log.With(logger, "fn", "CgnetService"),
		metrics: make(map[uint32]metrics.Counter),
		cache:   make(map[uint32]uint64),
	}
}

type cgService struct {
	cgroup string
	module *elf.Module

	logger log.Logger

	metrics map[uint32]metrics.Counter
	cache   map[uint32]uint64
}

func (s *cgService) AddCounterMetric(key uint32, c metrics.Counter) {
	s.metrics[key] = c
	s.cache[key] = 0
}

func (s *cgService) Attach(ctx context.Context, cgroupPath string) error {
	if cgroupPath == "" {
		return errors.New("cgroup path was \"\"")
	}
	s.cgroup = cgroupPath

	mod, err := bpf.Attach(cgroupPath)
	if err != nil {
		return err
	}
	s.module = mod

	level.Debug(s.logger).Log("msg", "Attaching BPF module successful")

	// init bpf map
	for k := range s.metrics {
		if err := bpf.UpdateKey(s.module, k, 0); err != nil {
			return errors.Wrap(err, "Failed to initialize map at key \"packets\"")
		}
	}
	level.Debug(s.logger).Log("msg", "Initializing BPF map successful")

	go s.run(ctx)
	level.Debug(s.logger).Log("msg", "Spawned BPF lookup worker")

	return nil
}

func (s *cgService) Detach(ctx context.Context) {
	if err := s.module.Close(); err != nil {
		level.Error(s.logger).Log("msg", "Error closing BPF module", "err", err)
		return
	}

	level.Debug(s.logger).Log("msg", "Detaching BPF module successful")
}

func (s *cgService) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case <-time.After(lookupPeriod):
			level.Debug(s.logger).Log("msg", "Looking up map keys")
			for k := range s.metrics {
				s.process(k)
			}
		}
	}
}

func (s *cgService) process(key uint32) {
	var logger = log.With(s.logger, "key", key)

	value, err := bpf.LookupKey(s.module, key)
	if err != nil {
		level.Warn(logger).Log("cgroup", s.cgroup, "msg", "Lookup failed", "err", err)
	}
	level.Debug(logger).Log("msg", "Read from map", "value", value)

	delta := float64((value - s.cache[key]))
	level.Debug(logger).Log("msg", "Calculate delta", "delta", delta)

	s.metrics[key].Add(delta)
	s.cache[key] = value
}
