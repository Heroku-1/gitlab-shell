package customaction

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/gitlab-shell/internal/command/readwriter"
	"gitlab.com/gitlab-org/gitlab-shell/internal/config"
	"gitlab.com/gitlab-org/gitlab-shell/internal/gitlabnet/accessverifier"
	"gitlab.com/gitlab-org/gitlab-shell/internal/gitlabnet/testserver"
)

func TestExecuteEOFSent(t *testing.T) {
	who := "key-1"

	requests := []testserver.TestRequestHandler{
		{
			Path: "/geo/proxy/info_refs_receive_pack",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				b, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				var request *Request
				require.NoError(t, json.Unmarshal(b, &request))

				require.Equal(t, request.Data.UserId, who)
				require.Empty(t, request.Output)

				err = json.NewEncoder(w).Encode(Response{Result: []byte("custom")})
				require.NoError(t, err)
			},
		},
		{
			Path: "/geo/proxy/receive_pack",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				b, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				var request *Request
				require.NoError(t, json.Unmarshal(b, &request))

				require.Equal(t, request.Data.UserId, who)
				require.Equal(t, "input", string(request.Output))

				err = json.NewEncoder(w).Encode(Response{Result: []byte("output")})
				require.NoError(t, err)
			},
		},
	}

	url, cleanup := testserver.StartSocketHttpServer(t, requests)
	defer cleanup()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	input := bytes.NewBufferString("input")

	response := &accessverifier.Response{
		Who: who,
		Payload: accessverifier.CustomPayload{
			Action: "geo_proxy_to_primary",
			Data: accessverifier.CustomPayloadData{
				ApiEndpoints: []string{"/geo/proxy/info_refs_receive_pack", "/geo/proxy/receive_pack"},
				Username:     "custom",
				PrimaryRepo:  "https://repo/path",
			},
		},
	}

	cmd := &Command{
		Config:     &config.Config{GitlabUrl: url},
		ReadWriter: &readwriter.ReadWriter{ErrOut: errBuf, Out: outBuf, In: input},
		EOFSent:    true,
	}

	require.NoError(t, cmd.Execute(response))

	// expect printing of info message, "custom" string from the first request
	// and "output" string from the second request
	require.Equal(t, "customoutput", outBuf.String())
}

func TestExecuteNoEOFSent(t *testing.T) {
	who := "key-1"

	requests := []testserver.TestRequestHandler{
		{
			Path: "/geo/proxy/info_refs_upload_pack",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				b, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				var request *Request
				require.NoError(t, json.Unmarshal(b, &request))

				require.Equal(t, request.Data.UserId, who)
				require.Empty(t, request.Output)

				err = json.NewEncoder(w).Encode(Response{Result: []byte("custom")})
				require.NoError(t, err)
			},
		},
		{
			Path: "/geo/proxy/upload_pack",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				b, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				var request *Request
				require.NoError(t, json.Unmarshal(b, &request))

				require.Equal(t, request.Data.UserId, who)
				require.Equal(t, "0032want 343d70886785dc1f98aaf70f3b4ca87c93a5d0dd\n", string(request.Output))

				err = json.NewEncoder(w).Encode(Response{Result: []byte("output")})
				require.NoError(t, err)
			},
		},
	}

	url, cleanup := testserver.StartSocketHttpServer(t, requests)
	defer cleanup()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	input := bytes.NewBufferString("0032want 343d70886785dc1f98aaf70f3b4ca87c93a5d0dd\n")

	response := &accessverifier.Response{
		Who: who,
		Payload: accessverifier.CustomPayload{
			Action: "geo_proxy_to_primary",
			Data: accessverifier.CustomPayloadData{
				ApiEndpoints: []string{"/geo/proxy/info_refs_upload_pack", "/geo/proxy/upload_pack"},
				Username:     "custom",
				PrimaryRepo:  "https://repo/path",
			},
		},
	}

	cmd := &Command{
		Config:     &config.Config{GitlabUrl: url},
		ReadWriter: &readwriter.ReadWriter{ErrOut: errBuf, Out: outBuf, In: input},
		EOFSent:    false,
	}

	require.NoError(t, cmd.Execute(response))

	// expect printing of info message, "custom" string from the first request
	// and "output" string from the second request
	require.Equal(t, "customoutput", outBuf.String())
}
