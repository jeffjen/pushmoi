package push

import (
	"github.com/jeffjen/pushmoi/oauth2"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"

	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"io"
	"os"
)

const (
	PUSH_NOTE_TYPE = "note"
	PUSH_LINK_TYPE = "link"
	PUSH_FILE_TYPE = "file"
)

type Push struct {
	// Targeting for push
	Iden       string `json:"device_iden,omitempty"`
	Email      string `json:"email,omitempty"`
	ChannelTag string `json:"channel_tag,omitempty"`
	ClientIden string `json:"client_iden,omitempty"`

	// Payload for the push
	Kind     string `json:"type"`
	Title    string `json:"title,omitempty"`
	Body     string `json:"body,omitempty"`
	Url      string `json:"url,omitempty"`
	FileName string `json:"file_name,omitempty"`
	FileType string `json:"file_type,omitempty"`
	FileUrl  string `json:"file_url,omitempty"`

	// Push initiator identity
	SrcIden string `json:"source_device_iden,omitempty"`

	// Unique identifier
	Guid string `json:"guid,omitempty"`
}

func NewPush(kind, title string) *Push {
	return &Push{Kind: kind, Title: title}
}

func (p *Push) Send(ctx context.Context) error {
	if p.Kind != PUSH_NOTE_TYPE && p.Kind != PUSH_LINK_TYPE && p.Kind != PUSH_FILE_TYPE {
		return errors.New("Reject push invalid push payload type")
	}

	var (
		cli = new(http.Client)
		buf = new(bytes.Buffer)
	)

	err := json.NewEncoder(buf).Encode(p)
	if err != nil {
		return errors.New("Malformed Pushbullet payload")
	}

	req, err := http.NewRequest("POST", "https://api.pushbullet.com/v2/pushes", buf)
	if err != nil {
		return errors.New("Failed to send push")
	}
	req.Header.Add("Access-Token", oauth2.Pushbullet.AccessToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := ctxhttp.Do(ctx, cli, req)
	if err != nil {
		return errors.New("Failed to send push")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(os.Stdout, resp.Body)
		return fmt.Errorf("%d", resp.StatusCode)
	}

	return nil
}
