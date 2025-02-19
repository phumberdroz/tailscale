// Copyright (c) 2020 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package monitor

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"tailscale.com/types/logger"
)

// unspecifiedMessage is a minimal message implementation that should not
// be ignored. In general, OS-specific implementations should use better
// types and avoid this if they can.
type unspecifiedMessage struct{}

func (unspecifiedMessage) ignore() bool { return false }

// devdConn implements osMon using devd(8).
type devdConn struct {
	conn net.Conn
}

func newOSMon(logf logger.Logf, m *Mon) (osMon, error) {
	conn, err := net.Dial("unixpacket", "/var/run/devd.seqpacket.pipe")
	if err != nil {
		logf("devd dial error: %v, falling back to polling method", err)
		return newPollingMon(logf, m)
	}
	return &devdConn{conn}, nil
}

func (c *devdConn) Close() error {
	return c.conn.Close()
}

func (c *devdConn) Receive() (message, error) {
	for {
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("reading devd socket: %v", err)
		}
		// Only return messages related to the network subsystem.
		if !strings.Contains(msg, "system=IFNET") {
			continue
		}
		// TODO: this is where the devd-specific message would
		// get converted into a "standard" event message and returned.
		return unspecifiedMessage{}, nil
	}
}
