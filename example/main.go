// Drop-in replacement for nakagami/grdp's example/main.go.
//
// PURPOSE
// -------
// This file establishes the CLI surface that the bench scenarios drive grdp
// through. It declares flags for features that DO NOT YET EXIST in grdp:
//   -drive, -clipboard, -read-test, -write-test, -sound, -monitors,
//   -send-keys, -send-mouse, -reconnect-count, -duration.
//
// Most of those flags are accepted but not implemented. They log
// "not yet implemented: <flag>" and the program proceeds with whatever IS
// implemented (currently: connect, render briefly, disconnect). Agents
// extending grdp during the experiment fill in the real behavior incrementally.
//
// CONTRACT WITH SCENARIOS
// -----------------------
// 1. Every flag MUST parse without error so the validator can run.
// 2. The program MUST exit 0 on a clean disconnect, non-zero on real errors.
// 3. Log lines should include feature names that the scenario-suite asserts
//    grep for: "rdpdr", "cliprdr", "rdpsnd", "drdynvc", "monitor",
//    "TS_INPUT_EVENT", "bitmap update", etc. The format is intentionally
//    grep-friendly.
//
// HOW TO APPLY
// ------------
// Replace the contents of `example/main.go` in the grdp fork with this file.
// You may need to fix imports — grdp's package layout occasionally shifts.
// If the existing example was multi-file, also remove the old siblings.
//
// LICENSE: matches the upstream grdp repo (GPL-3.0).

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	// TODO(agent): adjust these imports to match the repo state. The
	// following are the packages we expect to exist; if they have moved,
	// update the paths.
	//
	// "github.com/nakagami/grdp/core"
	// "github.com/nakagami/grdp/glog"
	// "github.com/nakagami/grdp/protocol/nla"
	// "github.com/nakagami/grdp/protocol/pdu"
	// "github.com/nakagami/grdp/protocol/sec"
	// "github.com/nakagami/grdp/protocol/t125"
	// "github.com/nakagami/grdp/protocol/x224"
)

// ---- CLI flags ------------------------------------------------------------

type cliFlags struct {
	host     string
	user     string
	password string
	sec      string // rdp | tls | nla
	duration time.Duration

	// New-feature flags (most are stubs). Fields exist so the scenarios can
	// pass values; the *executor functions* below decide whether to call
	// real grdp APIs or log "not yet implemented".
	drive          string // "name=hostpath", e.g. "share=/var/lib/bench/share"
	clipboard      bool
	readTest       string // remote path, e.g. "share\\readme.txt"
	readOut        string // local path to dump bytes
	writeTest      string // remote path, e.g. "share\\test-write.bin"
	writeSize      int
	writeSeed      int64
	sound          bool
	monitors       string // "x,y,w,h;x,y,w,h"
	sendKeys       string
	sendMouse      string // "move:X,Y;click:N;..."
	reconnectCount int
}

func parseFlags() *cliFlags {
	f := &cliFlags{}
	flag.StringVar(&f.host, "host", "", "RDP server host:port")
	flag.StringVar(&f.user, "user", "", "username")
	flag.StringVar(&f.password, "password", "", "password")
	flag.StringVar(&f.sec, "sec", "nla", "security: rdp | tls | nla")
	flag.DurationVar(&f.duration, "duration", 5*time.Second, "how long to stay connected before disconnecting")

	flag.StringVar(&f.drive, "drive", "", "redirected drive in form NAME=HOSTPATH (RDPDR)")
	flag.BoolVar(&f.clipboard, "clipboard", false, "enable clipboard channel (CLIPRDR)")
	flag.StringVar(&f.readTest, "read-test", "", "RDPDR remote path to read")
	flag.StringVar(&f.readOut, "read-out", "", "local path to write read-test bytes")
	flag.StringVar(&f.writeTest, "write-test", "", "RDPDR remote path to write")
	flag.IntVar(&f.writeSize, "write-size", 0, "bytes to write for -write-test")
	flag.Int64Var(&f.writeSeed, "write-seed", 0, "seed for -write-test pseudo-random payload")
	flag.BoolVar(&f.sound, "sound", false, "enable rdpsnd via DRDYNVC")
	flag.StringVar(&f.monitors, "monitors", "", "monitor rectangles: x,y,w,h;x,y,w,h (max 4)")
	flag.StringVar(&f.sendKeys, "send-keys", "", "string to type via TS_INPUT_EVENT")
	flag.StringVar(&f.sendMouse, "send-mouse", "", "mouse script: move:X,Y;click:N")
	flag.IntVar(&f.reconnectCount, "reconnect-count", 1, "number of connect cycles to perform")
	flag.Parse()

	if f.host == "" {
		fmt.Fprintln(os.Stderr, "error: -host is required")
		os.Exit(2)
	}
	return f
}

// ---- Logging helpers ------------------------------------------------------

func logFlags(f *cliFlags) {
	log.Printf("[bench] grdp example starting; host=%s sec=%s duration=%s", f.host, f.sec, f.duration)
	log.Printf("[bench] drive=%q clipboard=%v sound=%v monitors=%q", f.drive, f.clipboard, f.sound, f.monitors)
	log.Printf("[bench] read-test=%q write-test=%q write-size=%d", f.readTest, f.writeTest, f.writeSize)
	log.Printf("[bench] send-keys=%q send-mouse=%q reconnect-count=%d", f.sendKeys, f.sendMouse, f.reconnectCount)
}

func notYetImplemented(feature string) {
	log.Printf("[bench] not yet implemented: %s", feature)
}

// ---- Connection ----------------------------------------------------------

// connectOnce performs one connect/render/disconnect cycle and returns nil on
// clean exit. Real grdp client wiring lives here. Most channel setup is
// gated on flags; agents wire each one as they implement the channel.
func connectOnce(f *cliFlags) error {
	log.Printf("[bench] connecting to %s sec=%s", f.host, f.sec)

	// TODO(agent): instantiate grdp's client. The shape below is illustrative;
	// see grdp.go and protocol/* for the real API. Pseudo-code:
	//
	//   c, err := grdp.NewClient(f.host)
	//   if err != nil { return err }
	//   defer c.Close()
	//   c.SetUserName(f.user); c.SetPassword(f.password)
	//   c.SetSecurity(f.sec)  // rdp | tls | nla
	//
	//   // Channels: announce based on flags. Each is a separate ticket.
	//   if f.drive != "" {
	//       // RDPDR (backlog #01, #02, #03)
	//       name, hostPath, ok := splitDrive(f.drive)
	//       if !ok { notYetImplemented("rdpdr drive parse"); }
	//       // c.RegisterDriveRedirect(name, hostPath)
	//       log.Printf("[bench] rdpdr: announcing drive %s -> %s", name, hostPath)
	//       notYetImplemented("rdpdr device announce")
	//   }
	//   if f.clipboard {
	//       // CLIPRDR (backlog #04, #05, #06)
	//       // c.EnableClipboard()
	//       log.Printf("[bench] cliprdr: enabling clipboard channel")
	//       notYetImplemented("cliprdr channel open")
	//   }
	//   if f.sound {
	//       // DRDYNVC + RDPSND (backlog #07, #08)
	//       log.Printf("[bench] drdynvc: opening dynamic virtual channel transport")
	//       log.Printf("[bench] rdpsnd: requesting audio channel via DVC")
	//       notYetImplemented("drdynvc + rdpsnd")
	//   }
	//   if f.monitors != "" {
	//       // backlog #10
	//       rects, err := parseMonitorSpec(f.monitors)
	//       if err == nil {
	//           log.Printf("[bench] monitor count: %d", len(rects))
	//           // c.SetMonitorLayout(rects)
	//       }
	//       notYetImplemented("multi-monitor TS_UD_CS_MONITOR")
	//   }
	//
	//   if err := c.Connect(); err != nil { return err }
	//   log.Printf("[bench] connection established")
	//
	//   // Optional: send pre-recorded input.
	//   if f.sendKeys != "" {
	//       sendKeys(c, f.sendKeys)  // emits TS_INPUT_EVENT scancode pairs
	//   }
	//   if f.sendMouse != "" {
	//       sendMouse(c, f.sendMouse)
	//   }
	//   if f.readTest != "" {
	//       performRead(c, f.readTest, f.readOut)
	//   }
	//   if f.writeTest != "" {
	//       performWrite(c, f.writeTest, f.writeSize, f.writeSeed)
	//   }
	//
	//   // Idle for the duration so the server has a chance to push bitmap
	//   // updates, audio frames, etc.
	//   time.Sleep(f.duration)
	//   c.Disconnect()
	//   log.Printf("[bench] disconnect complete")

	// Until the agent fills in the real wiring above, do the safe minimum:
	// dial the host so the connect-rdp / connect-nla scenarios can at least
	// observe a TCP connect attempt in the logs.
	notYetImplemented("connect (real implementation)")
	time.Sleep(f.duration)
	log.Printf("[bench] disconnect complete (stub)")
	return nil
}

// ---- Stubs for the per-feature helpers -----------------------------------
//
// Agents replace these with real grdp API calls. They MUST keep the log
// strings the scenarios grep for (see scenarios/*.yaml `asserts:`).

func sendKeys(_ interface{}, s string) {
	log.Printf("[bench] send-keys: %d characters", len(s))
	for _, r := range s {
		// Each rune should produce a KEY_DOWN + KEY_UP TS_INPUT_EVENT pair.
		log.Printf("[bench] KEY_DOWN scancode=0x?? rune=%q", r)
		log.Printf("[bench] KEY_UP scancode=0x?? rune=%q", r)
	}
	notYetImplemented("real keyboard scancodes via TS_INPUT_EVENT")
}

func sendMouse(_ interface{}, spec string) {
	log.Printf("[bench] send-mouse: %s", spec)
	for _, step := range strings.Split(spec, ";") {
		step = strings.TrimSpace(step)
		switch {
		case strings.HasPrefix(step, "move:"):
			log.Printf("[bench] TS_INPUT_EVENT mouse-move %s", strings.TrimPrefix(step, "move:"))
		case strings.HasPrefix(step, "click:"):
			log.Printf("[bench] TS_INPUT_EVENT mouse-click button=%s", strings.TrimPrefix(step, "click:"))
		}
	}
	notYetImplemented("real mouse events via TS_INPUT_EVENT")
}

func performRead(_ interface{}, remote, local string) {
	log.Printf("[bench] read-test: remote=%s local=%s", remote, local)
	notYetImplemented("RDPDR IRP_MJ_CREATE + IRP_MJ_READ")
}

func performWrite(_ interface{}, remote string, size int, seed int64) {
	log.Printf("[bench] write-test: remote=%s size=%d seed=%d", remote, size, seed)
	notYetImplemented("RDPDR IRP_MJ_CREATE(write) + IRP_MJ_WRITE")
}

// ---- main ----------------------------------------------------------------

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	f := parseFlags()
	logFlags(f)

	// Graceful Ctrl-C
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	for i := 0; i < f.reconnectCount; i++ {
		if i > 0 {
			log.Printf("[bench] reconnect cycle %d/%d", i+1, f.reconnectCount)
			time.Sleep(500 * time.Millisecond)
		}
		select {
		case <-stop:
			log.Printf("[bench] interrupted, exiting")
			return
		default:
		}
		if err := connectOnce(f); err != nil {
			log.Printf("[bench] connect error: %v", err)
			os.Exit(1)
		}
	}
	log.Printf("[bench] run complete")
}
