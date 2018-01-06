// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build nacl solaris

package net

import "syscall"

func maxListenerBacklog() int {
	// TODO: Implement this id:1349 gh:1357
	// NOTE: Never return a number bigger than 1<<16 - 1. See issue 5030. id:1312 gh:1320
	return syscall.SOMAXCONN
}
