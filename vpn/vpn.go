// Package vpn has been created based on the following client:
// https://github.com/cghdev/gotunl.
package vpn

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

const (
	profileURL     = "http://unix/profile"
	authKeyPath    = "/var/run/pritunl.auth"
	unixSocketPath = "/var/run/pritunl.sock"
)

// ConnectionCredentials stores the credentials required to establish a new
// connection.
type ConnectionCredentials struct {
	ID  string
	Pin string
	OTP string
}

// Profile stores the Pritunl profile data.
type Profile struct {
	ID     string
	Name   string
	path   string
	config []byte
}

// Pritunl stores the data required to make requests to Pritunl.
type Pritunl struct {
	authKey     string
	profilePath string
}

// NewPritunl creates a new `Pritunl` client.
func NewPritunl() (*Pritunl, error) {
	p, err := &Pritunl{}, error(nil)

	p.authKey, err = getAuthKey()
	if err != nil {
		return nil, errors.Wrap(err, "error getting auth key")
	}
	p.profilePath, err = getProfilePath()
	if err != nil {
		return nil, errors.Wrap(err, "error getting profile path")
	}
	return p, nil
}

// ListProfiles lists Pritunl profiles.
func (p *Pritunl) ListProfiles() ([]Profile, error) {
	profiles := []Profile{}

	files, _ := filepath.Glob(p.profilePath + "/*.conf")
	for _, file := range files {
		id := strings.Split(filepath.Base(file), ".")[0]
		config, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, errors.Wrap(err, "error reading conf file")
		}

		name := gjson.GetBytes(config, "name").String()
		if name == "" {
			user := gjson.GetBytes(config, "user").String()
			server := gjson.GetBytes(config, "server").String()
			name = fmt.Sprintf("%s (%s)", user, server)
		}
		profiles = append(profiles, Profile{
			ID:     id,
			Name:   name,
			path:   file,
			config: config,
		})
	}

	return profiles, nil
}

// IsConnected checks if connection is established.
func (p *Pritunl) IsConnected(id string) (bool, error) {
	body, err := p.makeRequest(http.MethodGet, nil)
	if err != nil {
		return false, errors.Wrap(err, "error making request")
	}

	statuses := map[string]json.RawMessage{}
	if err = json.Unmarshal(body, &statuses); err != nil {
		return false, errors.Wrap(err, "error decoding response body")
	}

	status, ok := statuses[id]
	if !ok {
		return false, nil
	}

	if gjson.GetBytes(status, "status").String() != "connected" {
		return false, nil
	}

	return true, nil
}

// Connect sends a connect request to Pritunl.
func (p *Pritunl) Connect(creds ConnectionCredentials) error {
	profiles, err := p.ListProfiles()
	if err != nil {
		return errors.Wrap(err, "error listing profiles")
	}

	var profile *Profile
	for _, prof := range profiles {
		if prof.ID == creds.ID {
			profile = &prof
		}
	}
	if profile == nil {
		return fmt.Errorf("profile not found: %s", creds.ID)
	}

	if mode := gjson.GetBytes(profile.config, "password_mode").String(); mode != "otp_pin" {
		return fmt.Errorf("unsupported password mode: %s", mode)
	}

	ovpnFile := strings.Replace(profile.path, creds.ID+".conf", creds.ID+".ovpn", 1)
	ovpn, err := ioutil.ReadFile(ovpnFile)
	if err != nil {
		return errors.Wrap(err, "error reading ovpn file")
	}

	key := []byte{}
	if runtime.GOOS == "darwin" {
		command := "security find-generic-password -w -s pritunl -a " + creds.ID
		output, err := exec.Command("bash", "-c", command).Output()
		if err != nil {
			return errors.Wrap(err, "error executing command")
		}

		key = make([]byte, base64.StdEncoding.DecodedLen(len(output)))
		if _, err = base64.StdEncoding.Decode(key, output); err != nil {
			return errors.Wrap(err, "error decoding command output")
		}
	}

	serverPublicKeyParts := []string{}
	for _, part := range gjson.GetBytes(profile.config, "server_public_key").Array() {
		serverPublicKeyParts = append(serverPublicKeyParts, part.String())
	}

	// Send the same data as JS would send:
	// https://github.com/pritunl/pritunl-client-electron/blob/1.0.2395.64/client/www/js/service.js#L111.
	payloadData := map[string]interface{}{
		"id":                    creds.ID,
		"mode":                  "ovpn",
		"port_wg":               0,
		"org_id":                gjson.GetBytes(profile.config, "organization_id").String(),
		"user_id":               gjson.GetBytes(profile.config, "user_id").String(),
		"server_id":             gjson.GetBytes(profile.config, "server_id").String(),
		"sync_token":            gjson.GetBytes(profile.config, "sync_token").String(),
		"sync_secret":           gjson.GetBytes(profile.config, "sync_secret").String(),
		"username":              "pritunl",
		"password":              creds.Pin + creds.OTP,
		"server_public_key":     strings.Join(serverPublicKeyParts, "\n"),
		"server_box_public_key": gjson.GetBytes(profile.config, "server_box_public_key").String(),
		"token_ttl":             gjson.GetBytes(profile.config, "token_ttl").Int(),
		"reconnect":             true,
		"timeout":               true,
		"data":                  fmt.Sprintf("%s\n%s", ovpn, key),
	}

	payload, err := json.Marshal(payloadData)
	if err != nil {
		return errors.Wrap(err, "error creating payload")
	}

	_, err = p.makeRequest(http.MethodPost, payload)
	return errors.Wrap(err, "error making request")
}

// Disconnect sends a disconnect request to Pritunl.
func (p *Pritunl) Disconnect(id string) error {
	payloadData := map[string]interface{}{
		"id": id,
	}

	payload, err := json.Marshal(payloadData)
	if err != nil {
		return errors.Wrap(err, "error creating payload")
	}

	_, err = p.makeRequest(http.MethodDelete, payload)
	return errors.Wrap(err, "error making request")
}

// makeRequest makes request to Pritunl client.
func (p *Pritunl) makeRequest(verb string, payload []byte) ([]byte, error) {
	req, err := http.NewRequest(verb, profileURL, bytes.NewBuffer(payload))
	if err != nil {
		if err != nil {
			return nil, errors.Wrap(err, "error creating request")
		}
	}

	req.Header.Set("Auth-Key", p.authKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "close")
	req.Header.Set("User-Agent", "pritunl")

	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", unixSocketPath)
			},
		},
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error making request")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code: %d", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}
	return body, nil
}

// getAuthKey reads the auth key file and returns its contents.
func getAuthKey() (string, error) {
	if _, err := os.Stat(authKeyPath); !os.IsNotExist(err) {
		authKey, err := ioutil.ReadFile(authKeyPath)
		if err != nil {
			return "", errors.Wrap(err, "error reading file")
		}
		return string(authKey), nil
	}
	return "", errors.New("file not found")
}

// getProfilePath returns the path to Pritunl profile config files.
func getProfilePath() (string, error) {
	home, profilePath := os.Getenv("HOME"), ""
	switch runtime.GOOS {
	case "darwin":
		profilePath = home + "/Library/Application Support/pritunl/profiles"
	case "linux":
		profilePath = home + "/.config/pritunl/profiles"
	}
	if _, err := os.Stat(profilePath); !os.IsNotExist(err) {
		return profilePath, nil
	}
	return "", errors.New("path not found")
}
