package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	keyDir          = "/etc/cockpit-wg/keys"
	exchangePrivKey = keyDir + "/exchange.key"
	exchangePubKey  = keyDir + "/exchange.pub"
	signingPrivKey  = keyDir + "/signing.key"
	signingPubKey   = keyDir + "/signing.pub"
)

func ensureKeys() {
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		auditLog("EnsureKeys", json.RawMessage(fmt.Sprintf("{\"error\":\"%v\"}", err)), err)
		return
	}
	generated := false
	if _, err := os.Stat(exchangePrivKey); os.IsNotExist(err) {
		if err := exec.Command("age-keygen", "-o", exchangePrivKey).Run(); err == nil {
			os.Chmod(exchangePrivKey, 0600)
			os.Chown(exchangePrivKey, 0, 0)
			cmd := exec.Command("age-keygen", "-y", exchangePrivKey)
			if out, err := cmd.Output(); err == nil {
				os.WriteFile(exchangePubKey, out, 0600)
				os.Chown(exchangePubKey, 0, 0)
			}
			generated = true
		}
	}
	if _, err := os.Stat(signingPrivKey); os.IsNotExist(err) {
		if err := exec.Command("minisign", "-G", "-s", signingPrivKey, "-p", signingPubKey, "-n").Run(); err == nil {
			os.Chmod(signingPrivKey, 0600)
			os.Chown(signingPrivKey, 0, 0)
			os.Chmod(signingPubKey, 0600)
			os.Chown(signingPubKey, 0, 0)
			generated = true
		}
	}
	if generated {
		auditLog("GenerateKeys", json.RawMessage("{}"), nil)
	}
}

func getExchangeKey() (string, error) {
	b, err := os.ReadFile(exchangePubKey)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func rotateKeys() (string, error) {
	ts := time.Now().Format("20060102150405")
	for _, f := range []string{exchangePrivKey, exchangePubKey, signingPrivKey, signingPubKey} {
		if _, err := os.Stat(f); err == nil {
			os.Rename(f, f+"."+ts)
		}
	}
	oldPriv := exchangePrivKey + "." + ts
	if err := exec.Command("age-keygen", "-o", exchangePrivKey).Run(); err != nil {
		return "", err
	}
	os.Chmod(exchangePrivKey, 0600)
	os.Chown(exchangePrivKey, 0, 0)
	pubOut, err := exec.Command("age-keygen", "-y", exchangePrivKey).Output()
	if err != nil {
		return "", err
	}
	os.WriteFile(exchangePubKey, pubOut, 0600)
	os.Chown(exchangePubKey, 0, 0)
	if err := exec.Command("minisign", "-G", "-s", signingPrivKey, "-p", signingPubKey, "-n").Run(); err != nil {
		return "", err
	}
	os.Chmod(signingPrivKey, 0600)
	os.Chown(signingPrivKey, 0, 0)
	os.Chmod(signingPubKey, 0600)
	os.Chown(signingPubKey, 0, 0)
	reencryptInbox(oldPriv, strings.TrimSpace(string(pubOut)))
	auditLog("RotateKeys", json.RawMessage(fmt.Sprintf("{\"timestamp\":\"%s\"}", ts)), nil)
	return strings.TrimSpace(string(pubOut)), nil
}

func reencryptInbox(oldPriv, newPub string) {
	entries, err := os.ReadDir(inboxDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".wgx") {
			continue
		}
		path := filepath.Join(inboxDir, e.Name())
		tmp := path + ".tmp"
		dec := exec.Command("age", "-d", "-i", oldPriv, "-o", tmp, path)
		if err := dec.Run(); err != nil {
			os.Remove(tmp)
			continue
		}
		enc := exec.Command("age", "-r", newPub, "-o", path, tmp)
		enc.Run()
		os.Remove(tmp)
	}
}

func getSigningFingerprint() (string, error) {
	out, err := exec.Command("minisign", "-F", "-p", signingPubKey).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
