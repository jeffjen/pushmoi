package oauth2

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	path "path/filepath"
	"strings"
)

const (
	PUSH_BULLET_CONFIG = "~/.pushmoi/pushbullet.json"
)

type Device struct {
	Iden           string  `json:"iden"`
	Active         bool    `json:"active"`
	Created        float64 `json:"created"`
	Modified       float64 `json:"modififed"`
	Icon           string  `json:"icon"`
	Nickname       string  `json:"nickname"`
	IsGenerated    bool    `json:"generated_nickname"`
	Manufacturer   string  `json:"manufacturer"`
	Model          string  `json:"model"`
	Version        int     `json:"app_version"`
	Fingerprint    string  `json:"fingerprint"`
	KeyFingerprint string  `json:"key_fingerprint"`
	PushToken      string  `json:"push_token"`
	HasSms         bool    `json:"has_sms"`
}

type Devices struct {
	Devices []Device `json:"devices"`
}

func (d *Devices) Get() error {
	cli := new(http.Client)

	req, err := http.NewRequest("GET", "https://api.pushbullet.com/v2/devices", nil)
	if err != nil {
		return err
	}
	req.Header.Add("Access-Token", PushBullet.AccessToken)

	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(d)
	if err != nil {
		return err
	}

	return nil
}

type User struct {
	Iden            string  `json:"iden"`
	Created         float64 `json:"created"`
	Modified        float64 `json:"modififed"`
	Email           string  `json:"email"`
	EmailNormalized string  `json:"email_normalized"`
	Name            string  `json:"name"`
	ImageURL        string  `json:"image_url"`
	MaxUploadSize   int64   `json:"max_upload_size"`
	ReferredCount   int64   `json:"referred_count"`
	ReferredIden    string  `json:"referrer_iden"`
}

func (u *User) Get() error {
	cli := new(http.Client)

	req, err := http.NewRequest("GET", "https://api.pushbullet.com/v2/users/me", nil)
	if err != nil {
		return err
	}
	req.Header.Add("Access-Token", PushBullet.AccessToken)

	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(u)
}

type PushBulletConfig struct {
	*Devices
	*User

	AccessToken string `json:"access_token"`
}

func NewConfig() *PushBulletConfig {
	p := new(PushBulletConfig)
	p.Devices = new(Devices)
	p.User = new(User)
	return p
}

func (push *PushBulletConfig) Sync() error {
	// Sync current user profile
	if err := push.User.Get(); err != nil {
		return err
	}

	// Sync user registerd devices
	if err := push.Devices.Get(); err != nil {
		return err
	}

	return nil
}

func (push *PushBulletConfig) Load() error {
	conf, err := getConfigPath()
	if err != nil {
		return err
	}
	origin, err := os.OpenFile(conf, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer origin.Close()
	if err = json.NewDecoder(origin).Decode(push); err == io.EOF {
		return nil
	} else {
		return err
	}
}

func (push *PushBulletConfig) Dump() error {
	conf, err := getConfigPath()
	if err != nil {
		return err
	}
	origin, err := os.OpenFile(conf, os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer origin.Close()
	return json.NewEncoder(origin).Encode(push)
}

func getConfigPath() (string, error) {
	conf := strings.Replace(PUSH_BULLET_CONFIG, "~", os.Getenv("HOME"), 1)
	confdir := path.Dir(conf)
	if _, err := os.Stat(confdir); err != nil {
		if os.IsNotExist(err) {
			return conf, os.MkdirAll(confdir, 0700)
		} else {
			return "", err
		}
	} else {
		return conf, nil
	}
}
