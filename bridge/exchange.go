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
	"time"

	"github.com/coreos/go-systemd/v22/journal"
	"github.com/fsnotify/fsnotify"
)

const (
	inboxDir      = "/var/lib/cockpit-wg/inbox"
	trustedPubKey = "/var/lib/cockpit-wg/signing.pub"
	pendingDir    = "/var/lib/cockpit-wg/pending"
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

// exportBundle creates an encrypted and signed bundle for the given interface
// and returns the path to the generated .wgx file. The recipient parameter
// expects the target node's public exchange key.
func exportBundle(iface, recipient string) (string, error) {
	cfgPath := filepath.Join("/etc/wireguard", iface+".conf")
	cfg, err := os.ReadFile(cfgPath)
	if err != nil {
		auditExchange("export", iface, "", err)
		return "", err
	}
	sum := sha256.Sum256(cfg)
	manifest := Manifest{Interface: iface, Version: 1, Checksum: hex.EncodeToString(sum[:])}
	tmp, err := os.CreateTemp("", iface+"-*.tar")
	if err != nil {
		auditExchange("export", iface, hex.EncodeToString(sum[:]), err)
		return "", err
	}
	defer os.Remove(tmp.Name())
	tw := tar.NewWriter(tmp)
	manBytes, _ := json.Marshal(manifest)
	tw.WriteHeader(&tar.Header{Name: "manifest.json", Mode: 0600, Size: int64(len(manBytes))})
	tw.Write(manBytes)
	tw.WriteHeader(&tar.Header{Name: "config.conf", Mode: 0600, Size: int64(len(cfg))})
	tw.Write(cfg)
	tw.Close()
	tmp.Close()
	outName := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%d.wgx", iface, time.Now().UnixNano()))
	enc := exec.Command("age", "-r", recipient, "-o", outName, tmp.Name())
	if err := enc.Run(); err != nil {
		auditExchange("export", iface, hex.EncodeToString(sum[:]), err)
		os.Remove(outName)
		return "", err
	}
	sign := exec.Command("minisign", "-Sm", outName, "-s", signingPrivKey)
	if err := sign.Run(); err != nil {
		auditExchange("export", iface, hex.EncodeToString(sum[:]), err)
		os.Remove(outName)
		return "", err
	}
	auditExchange("export", iface, hex.EncodeToString(sum[:]), nil)
	return outName, nil
}

// listInboxBundles enumerates .wgx files in the inbox directory and returns a
// slice with basic verification results for each bundle.
func listInboxBundles() ([]map[string]interface{}, error) {
	entries, err := os.ReadDir(inboxDir)
	if err != nil {
		return nil, err
	}
	result := []map[string]interface{}{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".wgx") {
			continue
		}
		path := filepath.Join(inboxDir, e.Name())
		info := map[string]interface{}{"file": e.Name()}
		sig := path + ".minisig"
		if _, err := os.Stat(sig); err == nil {
			cmd := exec.Command("minisign", "-Vm", path, "-x", sig, "-P", trustedPubKey)
			if cmd.Run() == nil {
				info["signature"] = true
			} else {
				info["signature"] = false
			}
		} else {
			info["signature"] = false
		}
		decPath := path + ".tar"
		cmd := exec.Command("age", "-d", "-i", exchangePrivKey, "-o", decPath, path)
		if cmd.Run() == nil {
			info["recipient"] = true
			man, cfg, _, err := unpackBundle(decPath)
			if err == nil {
				sum := sha256.Sum256(cfg)
				info["checksum"] = strings.EqualFold(man.Checksum, hex.EncodeToString(sum[:]))
				info["interface"] = man.Interface
			} else {
				info["checksum"] = false
			}
		} else {
			info["recipient"] = false
		}
		os.Remove(decPath)
		result = append(result, info)
	}
	return result, nil
}

func auditExchange(action, iface, hash string, err error) {
	actor := os.Getenv("USER")
	fp, _ := getSigningFingerprint()
	fields := map[string]interface{}{"action": action, "actor": actor, "iface": iface, "hash": hash, "signing_fp": fp}
	if err != nil {
		fields["error"] = err.Error()
	}
	msgBytes, _ := json.Marshal(fields)
	journal.Send(string(msgBytes), journal.PriInfo, nil)
}
