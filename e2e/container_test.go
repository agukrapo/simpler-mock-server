package e2e

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/agukrapo/go-http-client/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func Test(t *testing.T) {
	ctx := context.Background()
	contReq := testcontainers.ContainerRequest{
		Image:        "simpler-mock-server:latest",
		ExposedPorts: []string{"4321/tcp"},
		WaitingFor:   wait.ForLog("Server started on :4321"),
	}
	sms, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: contReq,
		Started:          true,
	})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, sms.Terminate(ctx))
	}()

	endpoint, err := sms.Endpoint(ctx, "")
	require.NoError(t, err)

	t.Run("DELETE", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/api/people/a3b69b44-d562-11eb-b8bc-0242ac130003", endpoint)
		req, err := requests.New(url).Method(http.MethodDelete).Build(ctx)
		require.NoError(t, err)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusAccepted, res.StatusCode)
		assert.Equal(t, "text/html", res.Header.Get("Content-Type"))
		assert.Equal(t, string(b), "<html>\n<body>\n<h1>File deleted.</h1>\n</body>\n</html>")
	})

	t.Run("GET", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/health", endpoint)
		req, err := requests.New(url).Build(ctx)
		require.NoError(t, err)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))
		assert.Equal(t, string(b), "simpler-mock-server UP")
	})

	t.Run("PATCH", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/api/people/a3b69b44-d562-11eb-b8bc-0242ac130003", endpoint)
		req, err := requests.New(url).Method(http.MethodPatch).Build(ctx)
		require.NoError(t, err)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "application/xml", res.Header.Get("Content-Type"))
		assert.Equal(t, string(b), "<people>ok</people>")
	})

	t.Run("POST", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/api/people/a3b69b44-d562-11eb-b8bc-0242ac130003", endpoint)
		req, err := requests.New(url).Method(http.MethodPost).Build(ctx)
		require.NoError(t, err)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusTeapot, res.StatusCode)
		assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
		assert.Equal(t, string(b), "{\n  \"people\": \"418 I'm a teapot\"\n}")
	})
}
