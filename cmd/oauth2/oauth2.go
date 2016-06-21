package oauth2

import (
	"github.com/urfave/cli"
	"golang.org/x/net/context"

	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	OAUTH2_TIMEOUT_DURATION = 3 * time.Minute

	OAUTH2_WEB_URL = "localhost:8080/pushmoi/setup"
)

var (
	// PushBullet configuration
	PushBullet = NewConfig()
)

type oauth2Resp struct {
	token string
	err   error
}

func startOAuth2Workflow(ctx context.Context) <-chan oauth2Resp {
	var resp = make(chan oauth2Resp)
	go func() {
		// Create listener
		listen, err := net.Listen("tcp", ":8080")
		if err != nil {
			resp <- oauth2Resp{"", err}
			return
		}
		defer listen.Close()

		http.HandleFunc("/pushmoi/setup", func(w http.ResponseWriter, r *http.Request) {
			const (
				OAuth2ClientID     = "xroG8xHuOMNmYSfhBJjLw01YSP2XQCLa"
				OAuth2RedirectURI  = "http://localhost:8080/pushmoi/respond"
				OAuth2ResponseType = "token"
			)

			target := &url.URL{Scheme: "https", Host: "www.pushbullet.com", Path: "authorize"}

			qs := target.Query()
			qs.Set("client_id", OAuth2ClientID)
			qs.Set("redirect_uri", OAuth2RedirectURI)
			qs.Set("response_type", OAuth2ResponseType)

			target.RawQuery = qs.Encode()

			w.Header().Set("Location", target.String())

			w.WriteHeader(http.StatusFound)
		})

		// Serve user OAuth2 client authorization page
		http.HandleFunc("/pushmoi/respond", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			io.WriteString(w, OAUTH2_WORKFLOW_HTML)
		})

		// Receiver for client authorization
		http.HandleFunc("/pushmoi/authroized", func(w http.ResponseWriter, r *http.Request) {
			payload := struct {
				Token string `json:"access_token"`
			}{}

			if r.Method != "POST" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			if json.NewDecoder(r.Body).Decode(&payload) != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			resp <- oauth2Resp{payload.Token, nil}
			io.WriteString(w, "ok")
		})

		server := new(http.Server)
		go server.Serve(listen)

		// If we reached here we were canceled
		<-ctx.Done()
	}()
	return resp
}

func continueSetup(token string) error {
	defer PushBullet.Dump()

	// Store access_token to Signin User
	PushBullet.AccessToken = token

	ctx, _ := context.WithTimeout(context.Background(), 1*time.Minute)

	// Sync user profile and registered devices
	if err := PushBullet.Sync(ctx); err != nil {
		return cli.NewExitError("Failed to sync PushBullet info", 3)
	} else {
		return nil
	}
}

func NewOAuth2Workflow() cli.Command {
	return cli.Command{
		Name:  "init",
		Usage: "Initialize pushmoi client with PushBullet",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "token", Usage: "Use this access_token to initialize"},
		},
		Action: func(c *cli.Context) error {
			token := c.String("token")
			if token == "" {
				ctx, _ := context.WithTimeout(context.Background(), OAUTH2_TIMEOUT_DURATION)
				srvctx, shutdown := context.WithCancel(ctx)
				access_token := startOAuth2Workflow(srvctx)
				fmt.Println("Please signin through the following URL:\n")
				fmt.Println("  ", OAUTH2_WEB_URL, "\n")
				defer shutdown()
				select {
				case <-ctx.Done():
					return cli.NewExitError("Operation failed to complete", 1)
				case resp := <-access_token:
					if resp.err != nil {
						return cli.NewExitError("Failed to obtain access token", 2)
					} else {
						return continueSetup(resp.token)
					}
				}
			} else {
				return continueSetup(token)
			}
		},
	}
}
