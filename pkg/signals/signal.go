/*
 * A lightweight Go-based controller that automatically exposes deployments using NodePort services
 * Designed for simplicity, reliability, and easy integration into Kubernetes workflows.
 *
 * Copyright (C) 2025 Abdul Saqib
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

// Package signals provides utilities for handling OS signals such as SIGINT and SIGTERM.
package signals

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"

	"k8s.io/klog/v2"
)

const signalChannelBuffer = 2

// SetupSignalHandler returns a context that is canceled on the first SIGINT or SIGTERM.
// A second signal will cancel the context again, allowing the caller to handle immediate shutdown.
func SetupSignalHandler() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, signalChannelBuffer)
	signal.Notify(sigChan, shutdownSignals()...)

	var handled int32
	go func() {
		for range sigChan {
			if !atomic.CompareAndSwapInt32(&handled, 0, 1) {
				cancel()
				klog.Info("received second shutdown signal, canceling again immediately...")
				return
			}
			cancel()
			klog.Info("received first shutdown signal, canceling context...")
		}
	}()
	return ctx
}

func shutdownSignals() []os.Signal {
	shutdownSignals := []os.Signal{os.Interrupt, syscall.SIGTERM}
	if runtime.GOOS == "windows" {
		shutdownSignals = []os.Signal{os.Interrupt}
	}
	return shutdownSignals
}
