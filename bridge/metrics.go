package main

import (
	"os"
	"strconv"
	"sync"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
)

type metricSample struct {
	Time time.Time
	Rx   float64
	Tx   float64
}

type ringBuffer struct {
	data   []metricSample
	next   int
	filled bool
}

func newRingBuffer(size int) *ringBuffer {
	return &ringBuffer{data: make([]metricSample, size)}
}

func (r *ringBuffer) add(s metricSample) {
	if len(r.data) == 0 {
		return
	}
	r.data[r.next] = s
	r.next++
	if r.next >= len(r.data) {
		r.next = 0
		r.filled = true
	}
}

func (r *ringBuffer) samples() []metricSample {
	if len(r.data) == 0 {
		return []metricSample{}
	}
	if !r.filled {
		return r.data[:r.next]
	}
	out := make([]metricSample, len(r.data))
	copy(out, r.data[r.next:])
	copy(out[len(r.data)-r.next:], r.data[:r.next])
	return out
}

type prevStats struct {
	rx int64
	tx int64
	t  time.Time
}

type metricsCollector struct {
	mu       sync.Mutex
	interval time.Duration
	buffers  map[string]*ringBuffer
	prev     map[string]prevStats
	client   *wgctrl.Client
}

var collector *metricsCollector

func initMetricsCollector() {
	interval := metricsInterval()
	c, err := newMetricsCollector(interval)
	if err != nil {
		return
	}
	collector = c
	go collector.run()
}

func metricsInterval() time.Duration {
	v := os.Getenv("WG_METRICS_INTERVAL")
	if v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			if i < 2 {
				i = 2
			} else if i > 10 {
				i = 10
			}
			return time.Duration(i) * time.Second
		}
	}
	return 2 * time.Second
}

func newMetricsCollector(interval time.Duration) (*metricsCollector, error) {
	client, err := wgctrl.New()
	if err != nil {
		return nil, err
	}
	return &metricsCollector{
		interval: interval,
		buffers:  make(map[string]*ringBuffer),
		prev:     make(map[string]prevStats),
		client:   client,
	}, nil
}

func (m *metricsCollector) run() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	for range ticker.C {
		m.collect()
	}
}

func (m *metricsCollector) collect() {
	devices, err := m.client.Devices()
	if err != nil {
		return
	}
	now := time.Now()
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, dev := range devices {
		name := dev.Name
		var rx, tx int64
		for _, p := range dev.Peers {
			rx += p.ReceiveBytes
			tx += p.TransmitBytes
		}
		prev, ok := m.prev[name]
		sample := metricSample{Time: now}
		if ok {
			dt := now.Sub(prev.t).Seconds()
			if dt > 0 {
				if rx >= prev.rx {
					sample.Rx = float64(rx-prev.rx) / dt
				}
				if tx >= prev.tx {
					sample.Tx = float64(tx-prev.tx) / dt
				}
			}
		}
		m.prev[name] = prevStats{rx: rx, tx: tx, t: now}
		buf, ok := m.buffers[name]
		if !ok {
			bufLen := int((5 * time.Minute) / m.interval)
			buf = newRingBuffer(bufLen)
			m.buffers[name] = buf
		}
		buf.add(sample)
	}
}

func (m *metricsCollector) getMetrics(name string) []metricSample {
	m.mu.Lock()
	defer m.mu.Unlock()
	buf, ok := m.buffers[name]
	if !ok {
		return []metricSample{}
	}
	samples := buf.samples()
	out := make([]metricSample, len(samples))
	copy(out, samples)
	return out
}

func getMetrics(name string) (interface{}, error) {
	if collector == nil {
		return nil, ErrMetricsUnavailable
	}
	samples := collector.getMetrics(name)
	times := make([]int64, len(samples))
	rx := make([]float64, len(samples))
	tx := make([]float64, len(samples))
	for i, s := range samples {
		times[i] = s.Time.Unix()
		rx[i] = s.Rx
		tx[i] = s.Tx
	}
	return map[string]interface{}{"timestamps": times, "rx": rx, "tx": tx}, nil
}
