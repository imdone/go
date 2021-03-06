// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package main

import (
	"bytes"
	"internal/trace"
	"runtime"
	rtrace "runtime/trace"
	"sync"
	"syscall"
	"testing"
	"time"
)

// TestGoroutineInSyscall tests threads for timer goroutines
// that preexisted when the tracing started were not counted
// as threads in syscall. See golang.org/issues/22574.
func TestGoroutineInSyscall(t *testing.T) {
	// Start one goroutine blocked in syscall.
	//
	// TODO: syscall.Pipe used to cause the goroutine to id:984 gh:992
	// remain blocked in syscall is not portable. Replace
	// it with a more portable way so this test can run
	// on non-unix architecture e.g. Windows.
	var p [2]int
	if err := syscall.Pipe(p[:]); err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	var wg sync.WaitGroup
	defer func() {
		syscall.Write(p[1], []byte("a"))
		wg.Wait()

		syscall.Close(p[0])
		syscall.Close(p[1])
	}()
	wg.Add(1)
	go func() {
		var tmp [1]byte
		syscall.Read(p[0], tmp[:])
		wg.Done()
	}()

	// Start multiple timer goroutines.
	allTimers := make([]*time.Timer, 2*runtime.GOMAXPROCS(0))
	defer func() {
		for _, timer := range allTimers {
			timer.Stop()
		}
	}()

	var timerSetup sync.WaitGroup
	for i := range allTimers {
		timerSetup.Add(1)
		go func(i int) {
			defer timerSetup.Done()
			allTimers[i] = time.AfterFunc(time.Hour, nil)
		}(i)
	}
	timerSetup.Wait()

	// Collect and parse trace.
	buf := new(bytes.Buffer)
	if err := rtrace.Start(buf); err != nil {
		t.Fatalf("failed to start tracing: %v", err)
	}
	rtrace.Stop()

	res, err := trace.Parse(buf, "")
	if err != nil {
		t.Fatalf("failed to parse trace: %v", err)
	}

	// Check only one thread for the pipe read goroutine is
	// considered in-syscall.
	viewerData, err := generateTrace(&traceParams{
		parsed:  res,
		endTime: int64(1<<63 - 1),
	})
	if err != nil {
		t.Fatalf("failed to generate ViewerData: %v", err)
	}
	for _, ev := range viewerData.Events {
		if ev.Name == "Threads" {
			arg := ev.Arg.(*threadCountersArg)
			if arg.InSyscall > 1 {
				t.Errorf("%d threads in syscall at time %v; want less than 1 thread in syscall", arg.InSyscall, ev.Time)
			}
		}
	}
}
