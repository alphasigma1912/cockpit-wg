package main

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/coreos/go-systemd/v22/journal"
	"github.com/fsnotify/fsnotify"
)

const (
	inboxDir        = "/var/lib/cockpit-wg/inbox"
	trustedPubKey   = "/var/lib/cockpit-wg/signing.pub"
	exchangePrivKey = "/var/lib/cockpit-wg/exchange.key"
	pendingDir      = "/var/lib/cockpit-wg/pending"
)

type Manifest struct {
	Interface string `json:"interface"`
	Version   int    `json:"version"`
	Checksum  string `json:"checksum"`
}

func watchInbox() {
	if err := os.MkdirAll(inboxDir, 0700); err != nil {
		journal.Send(fmt.Sprintf("{\"action\":\"inbox\",\"error\":\"%v\"}", err), journal.PriErr, nil)
		return
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		journal.Send(fmt.Sprintf("{\"action\":\"inbox\",\"error\":\"%v\"}", err), journal.PriErr, nil)
		return
	}
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create && strings.HasSuffix(event.Name, ".wgx") {
					go handleBundle(event.Name)
				}
			case err := <-watcher.Errors:
				journal.Send(fmt.Sprintf("{\"action\":\"inbox\",\"error\":\"%v\"}", err), journal.PriErr, nil)
			}
		}
	}()
	watcher.Add(inboxDir)
}

func handleBundle(path string) {
	sig := path + ".minisig"
	if _, err := os.Stat(sig); err != nil {
		journal.Send(fmt.Sprintf("{\"bundle\":\"%s\",\"error\":\"missing signature\"}", filepath.Base(path)), journal.PriErr, nil)
		return
	}
	cmd := exec.Command("minisign", "-Vm", path, "-x", sig, "-P", trustedPubKey)
	if err := cmd.Run(); err != nil {
		journal.Send(fmt.Sprintf("{\"bundle\":\"%s\",\"error\":\"signature verify failed\"}", filepath.Base(path)), journal.PriErr, nil)
		return
	}
	decPath := path + ".tar"
	cmd = exec.Command("age", "-d", "-i", exchangePrivKey, "-o", decPath, path)
	if err := cmd.Run(); err != nil {
		journal.Send(fmt.Sprintf("{\"bundle\":\"%s\",\"error\":\"decrypt failed\"}", filepath.Base(path)), journal.PriErr, nil)
		return
	}
	manifest, cfg, meta, err := unpackBundle(decPath)
	if err != nil {
		journal.Send(fmt.Sprintf("{\"bundle\":\"%s\",\"error\":\"%v\"}", filepath.Base(path), err), journal.PriErr, nil)
		return
	}
	sum := sha256.Sum256(cfg)
	if hex.EncodeToString(sum[:]) != strings.ToLower(manifest.Checksum) {
		journal.Send(fmt.Sprintf("{\"bundle\":\"%s\",\"error\":\"checksum mismatch\"}", filepath.Base(path)), journal.PriErr, nil)
		return
	}
	dest := filepath.Join(pendingDir, manifest.Interface)
	if err := os.MkdirAll(filepath.Join(dest, "meta"), 0700); err != nil {
		journal.Send(fmt.Sprintf("{\"bundle\":\"%s\",\"error\":\"%v\"}", filepath.Base(path), err), journal.PriErr, nil)
		return
	}
	if err := os.WriteFile(filepath.Join(dest, "config.conf"), cfg, 0600); err != nil {
		journal.Send(fmt.Sprintf("{\"bundle\":\"%s\",\"error\":\"%v\"}", filepath.Base(path), err), journal.PriErr, nil)
		return
	}
	for name, data := range meta {
		os.WriteFile(filepath.Join(dest, name), data, 0600)
	}
	journal.Send(fmt.Sprintf("{\"action\":\"bundle\",\"iface\":\"%s\",\"status\":\"ready\"}", manifest.Interface), journal.PriInfo, nil)
	os.Remove(path)
	os.Remove(sig)
	os.Remove(decPath)
}

func unpackBundle(tarPath string) (*Manifest, []byte, map[string][]byte, error) {
	f, err := os.Open(tarPath)
	if err != nil {
		return nil, nil, nil, err
	}
	defer f.Close()
	tr := tar.NewReader(f)
	var manifest Manifest
	var cfg []byte
	meta := make(map[string][]byte)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, nil, err
		}
		switch hdr.Name {
		case "manifest.json":
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, tr); err != nil {
				return nil, nil, nil, err
			}
			if err := json.Unmarshal(buf.Bytes(), &manifest); err != nil {
				return nil, nil, nil, err
			}
		case "config.conf":
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, tr); err != nil {
				return nil, nil, nil, err
			}
			cfg = buf.Bytes()
		default:
			if strings.HasPrefix(hdr.Name, "meta/") && !hdr.FileInfo().IsDir() {
				var buf bytes.Buffer
				if _, err := io.Copy(&buf, tr); err != nil {
					return nil, nil, nil, err
				}
				meta[hdr.Name] = buf.Bytes()
			}
		}
	}
	if manifest.Interface == "" || len(cfg) == 0 {
		return nil, nil, nil, fmt.Errorf("incomplete bundle")
	}
	return &manifest, cfg, meta, nil
}
