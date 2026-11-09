package main

import (
	"fmt"
	"io"
	"sync"
	"time"

	"omnitranscripts/engine"
)

// progressMeter renders concise, carriage-return-updated pipeline progress.
// The meter intentionally reports stages and elapsed time instead of inventing
// a percentage from subprocesses that do not expose a shared progress scale.
type progressMeter struct {
	writer io.Writer
	start  time.Time

	mu      sync.Mutex
	stage   engine.Stage
	frame   int
	stopped bool
	stop    chan struct{}
	done    chan struct{}
}

func newProgressMeter(writer io.Writer) *progressMeter {
	return &progressMeter{
		writer: writer,
		stage:  engine.StageDownload,
		stop:   make(chan struct{}),
		done:   make(chan struct{}),
	}
}

func (m *progressMeter) Start() {
	m.mu.Lock()
	m.start = time.Now()
	m.mu.Unlock()
	go m.loop()
}

func (m *progressMeter) Update(stage engine.Stage) {
	m.mu.Lock()
	m.stage = stage
	m.mu.Unlock()
}

func (m *progressMeter) Stop(success bool) {
	m.mu.Lock()
	if m.stopped {
		m.mu.Unlock()
		return
	}
	m.stopped = true
	m.mu.Unlock()
	close(m.stop)
	<-m.done

	m.mu.Lock()
	elapsed := time.Since(m.start).Round(time.Second)
	status := "failed"
	if success {
		status = "complete"
	}
	fmt.Fprintf(m.writer, "\rTranscribing [%s] %s in %s\n", status, stageLabel(m.stage), elapsed)
	m.mu.Unlock()
}

func (m *progressMeter) loop() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	defer close(m.done)
	for {
		select {
		case <-ticker.C:
			m.render()
		case <-m.stop:
			return
		}
	}
}

func (m *progressMeter) render() {
	m.mu.Lock()
	defer m.mu.Unlock()
	frames := []rune{'|', '/', '-', '\\'}
	frame := frames[m.frame%len(frames)]
	m.frame++
	fmt.Fprintf(m.writer, "\rTranscribing %c [%s] %s", frame, stageLabel(m.stage), time.Since(m.start).Round(time.Second))
}

func stageLabel(stage engine.Stage) string {
	switch stage {
	case engine.StageDownload:
		return "1/3 download"
	case engine.StageNormalize:
		return "2/3 normalize"
	case engine.StageTranscribe:
		return "3/3 transcribe"
	default:
		return string(stage)
	}
}
