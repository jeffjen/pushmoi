package oauth2

import (
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"

	"encoding/json"
	"errors"
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

type Devs struct {
	Devices []*Device `json:"devices"`
}

func (d *Devs) Get(ctx context.Context) error {
	cli := new(http.Client)

	req, err := http.NewRequest("GET", "https://api.pushbullet.com/v2/devices", nil)
	if err != nil {
		return err
	}
	req.Header.Add("Access-Token", Pushbullet.AccessToken)

	resp, err := ctxhttp.Do(ctx, cli, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Failed to obtain registered devices")
	}

	return json.NewDecoder(resp.Body).Decode(d)
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

func (u *User) Get(ctx context.Context) error {
	cli := new(http.Client)

	req, err := http.NewRequest("GET", "https://api.pushbullet.com/v2/users/me", nil)
	if err != nil {
		return err
	}
	req.Header.Add("Access-Token", Pushbullet.AccessToken)

	resp, err := ctxhttp.Do(ctx, cli, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Failed to obtain user info")
	}

	return json.NewDecoder(resp.Body).Decode(u)
}

type PushBulletConfig struct {
	*Devs
	*User

	AccessToken string `json:"access_token"`
}

func NewConfig() *PushBulletConfig {
	p := new(PushBulletConfig)
	p.Devs = new(Devs)
	p.User = new(User)
	return p
}

func (push *PushBulletConfig) Has(name string) *Device {
	for _, dev := range push.Devices {
		if dev.Nickname == name {
			return dev
		}
	}
	return nil
}

func (push *PushBulletConfig) Sync(ctx context.Context) error {
	// Sync current user profile
	if err := push.User.Get(ctx); err != nil {
		return err
	}

	// Sync user registerd devices
	if err := push.Devs.Get(ctx); err != nil {
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
