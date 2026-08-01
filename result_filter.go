package main

import (
	"fmt"
	"time"

	"github.com/faceair/clash-speedtest/speedtester"
)

type resultFilter struct {
	mode             speedtester.SpeedMode
	maxLatency       time.Duration
	maxPacketLoss    float64
	minDownloadSpeed float64
	minUploadSpeed   float64
	downloadSize     int
}

func newResultFilter(mode speedtester.SpeedMode) resultFilter {
	return resultFilter{
		mode:             mode,
		maxLatency:       *maxLatency,
		maxPacketLoss:    *maxPacketLoss,
		minDownloadSpeed: *minDownloadSpeed * 1024 * 1024,
		minUploadSpeed:   *minUploadSpeed * 1024 * 1024,
		downloadSize:     *downloadSize,
	}
}

func (f resultFilter) Match(result *speedtester.Result) bool {
	if result == nil {
		return false
	}
	// Failed probes report Latency=0 and PacketLoss=100; never export them.
	if result.Latency <= 0 {
		return false
	}
	if result.PacketLoss >= 100 {
		return false
	}
	if f.maxLatency > 0 && result.Latency > f.maxLatency {
		return false
	}
	if f.maxPacketLoss >= 0 && result.PacketLoss > f.maxPacketLoss {
		return false
	}
	// fast 模式不会测速，DownloadSpeed 为 0，此时只按延迟和丢包筛选。
	if !f.mode.IsFast() && f.downloadSize > 0 && f.minDownloadSpeed > 0 && result.DownloadSpeed < f.minDownloadSpeed {
		return false
	}
	if f.mode.UploadEnabled() && f.minUploadSpeed > 0 && result.UploadSpeed < f.minUploadSpeed {
		return false
	}
	if result.ProxyConfig == nil || result.ProxyConfig["name"] == nil || result.ProxyConfig["server"] == nil {
		return false
	}
	return true
}

type earlyStopper struct {
	limit  int
	filter resultFilter
	count  int
}

func newEarlyStopper(limit int, filter resultFilter) (*earlyStopper, error) {
	if limit < 0 {
		return nil, fmt.Errorf("early-stop must be greater than or equal to 0, got %d", limit)
	}
	return &earlyStopper{
		limit:  limit,
		filter: filter,
	}, nil
}

func (s *earlyStopper) ShouldContinue(result *speedtester.Result) bool {
	if s.limit == 0 {
		return true
	}
	if s.filter.Match(result) {
		s.count++
	}
	return s.count < s.limit
}
