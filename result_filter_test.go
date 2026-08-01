package main

import (
	"testing"
	"time"

	"github.com/faceair/clash-speedtest/speedtester"
)

func TestResultFilterMatch(t *testing.T) {
	validResult := func() *speedtester.Result {
		return &speedtester.Result{
			ProxyConfig:   map[string]any{"name": "proxy-a", "server": "1.1.1.1"},
			Latency:       50 * time.Millisecond,
			PacketLoss:    1,
			DownloadSpeed: 10 * 1024 * 1024,
			UploadSpeed:   5 * 1024 * 1024,
		}
	}

	baseFilter := resultFilter{
		mode:             speedtester.SpeedModeDownload,
		maxLatency:       100 * time.Millisecond,
		maxPacketLoss:    10,
		minDownloadSpeed: 5 * 1024 * 1024,
		minUploadSpeed:   2 * 1024 * 1024,
		downloadSize:     1024,
	}

	if !baseFilter.Match(validResult()) {
		t.Fatalf("expected valid download result to pass filters")
	}

	t.Run("filters by latency", func(t *testing.T) {
		result := validResult()
		result.Latency = 150 * time.Millisecond
		if baseFilter.Match(result) {
			t.Fatalf("expected high latency result to fail")
		}
	})

	t.Run("filters by packet loss", func(t *testing.T) {
		result := validResult()
		result.PacketLoss = 20
		if baseFilter.Match(result) {
			t.Fatalf("expected high packet loss result to fail")
		}
	})

	t.Run("filters by download speed in download mode", func(t *testing.T) {
		result := validResult()
		result.DownloadSpeed = 1024
		if baseFilter.Match(result) {
			t.Fatalf("expected slow download result to fail")
		}
	})

	t.Run("fast mode ignores download speed", func(t *testing.T) {
		filter := baseFilter
		filter.mode = speedtester.SpeedModeFast
		result := validResult()
		result.DownloadSpeed = 0
		if !filter.Match(result) {
			t.Fatalf("expected fast mode to ignore download speed")
		}
	})

	t.Run("full mode filters by upload speed", func(t *testing.T) {
		filter := baseFilter
		filter.mode = speedtester.SpeedModeFull
		result := validResult()
		result.UploadSpeed = 1024
		if filter.Match(result) {
			t.Fatalf("expected slow upload result to fail in full mode")
		}
	})

	t.Run("requires writable proxy config entry", func(t *testing.T) {
		result := validResult()
		delete(result.ProxyConfig, "server")
		if baseFilter.Match(result) {
			t.Fatalf("expected result without server to fail")
		}
		if baseFilter.Match(nil) {
			t.Fatalf("expected nil result to fail")
		}
	})

	t.Run("rejects failed latency probes", func(t *testing.T) {
		result := validResult()
		result.Latency = 0
		result.PacketLoss = 100
		if baseFilter.Match(result) {
			t.Fatalf("expected zero-latency failed probe to fail")
		}

		filter := baseFilter
		filter.maxPacketLoss = 100
		result = validResult()
		result.Latency = 0
		result.PacketLoss = 100
		if filter.Match(result) {
			t.Fatalf("expected 100%% packet loss to fail even when max-packet-loss is 100")
		}
	})
}

func TestEarlyStopperShouldContinue(t *testing.T) {
	filter := resultFilter{
		mode:             speedtester.SpeedModeDownload,
		maxLatency:       100 * time.Millisecond,
		maxPacketLoss:    10,
		minDownloadSpeed: 5 * 1024 * 1024,
		downloadSize:     1024,
	}
	matching := &speedtester.Result{
		ProxyConfig:   map[string]any{"name": "proxy-a", "server": "1.1.1.1"},
		Latency:       50 * time.Millisecond,
		PacketLoss:    0,
		DownloadSpeed: 6 * 1024 * 1024,
	}
	notMatching := &speedtester.Result{
		ProxyConfig:   map[string]any{"name": "proxy-b", "server": "2.2.2.2"},
		Latency:       50 * time.Millisecond,
		PacketLoss:    0,
		DownloadSpeed: 1024,
	}

	stopper, err := newEarlyStopper(2, filter)
	if err != nil {
		t.Fatalf("create early stopper failed: %s", err)
	}

	if !stopper.ShouldContinue(notMatching) {
		t.Fatalf("expected non-matching result not to trigger early stop")
	}
	if stopper.count != 0 {
		t.Fatalf("expected non-matching result not to increase count, got %d", stopper.count)
	}
	if !stopper.ShouldContinue(matching) {
		t.Fatalf("expected first matching result to continue")
	}
	if stopper.count != 1 {
		t.Fatalf("expected first matching result to increase count to 1, got %d", stopper.count)
	}
	if stopper.ShouldContinue(matching) {
		t.Fatalf("expected second matching result to stop")
	}
	if stopper.count != 2 {
		t.Fatalf("expected second matching result to increase count to 2, got %d", stopper.count)
	}

	disabled, err := newEarlyStopper(0, filter)
	if err != nil {
		t.Fatalf("create disabled early stopper failed: %s", err)
	}
	if !disabled.ShouldContinue(matching) || !disabled.ShouldContinue(matching) {
		t.Fatalf("expected disabled early stopper to always continue")
	}
	if disabled.count != 0 {
		t.Fatalf("expected disabled early stopper not to count matches, got %d", disabled.count)
	}

	if _, err := newEarlyStopper(-1, filter); err == nil {
		t.Fatalf("expected negative limit to fail")
	}
}
