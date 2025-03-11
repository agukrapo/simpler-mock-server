package e2e

import (
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
	contReq := testcontainers.ContainerRequest{
		Image:        "simpler-mock-server:latest",
		ExposedPorts: []string{"4321/tcp"},
		WaitingFor:   wait.ForLog("Server started on :4321"),
	}
	sms, err := testcontainers.GenericContainer(t.Context(), testcontainers.GenericContainerRequest{
		ContainerRequest: contReq,
		Started:          true,
	})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, sms.Terminate(t.Context()))
	}()

	endpoint, err := sms.Endpoint(t.Context(), "")
	require.NoError(t, err)

	t.Run("DELETE", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/api/people/a3b69b44-d562-11eb-b8bc-0242ac130003", endpoint)
		req, err := requests.New(url).Method(http.MethodDelete).Build(t.Context())
		require.NoError(t, err)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusAccepted, res.StatusCode)
		assert.Equal(t, "text/html", res.Header.Get("Content-Type"))
		assert.Equal(t, "<html>\n<body>\n<h1>File deleted.</h1>\n</body>\n</html>", string(b))
	})

	t.Run("GET", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/health", endpoint)
		req, err := requests.New(url).Build(t.Context())
		require.NoError(t, err)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))
		assert.Equal(t, "simpler-mock-server UP", string(b))
	})

	t.Run("PATCH", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/api/people/a3b69b44-d562-11eb-b8bc-0242ac130003", endpoint)
		req, err := requests.New(url).Method(http.MethodPatch).Build(t.Context())
		require.NoError(t, err)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "application/xml", res.Header.Get("Content-Type"))
		assert.Equal(t, "<people>ok</people>", string(b))
	})

	t.Run("POST", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/api/people/a3b69b44-d562-11eb-b8bc-0242ac130003", endpoint)
		req, err := requests.New(url).Method(http.MethodPost).Build(t.Context())
		require.NoError(t, err)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusTeapot, res.StatusCode)
		assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
		assert.JSONEq(t, "{\n  \"people\": \"418 I'm a teapot\"\n}", string(b))
	})

	t.Run("PUT", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/api/people", endpoint)
		req, err := requests.New(url).Method(http.MethodPut).Build(t.Context())
		require.NoError(t, err)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNoContent, res.StatusCode)
		assert.Equal(t, "text/csv", res.Header.Get("Content-Type"))
		assert.Empty(t, b)
	})

	t.Run("create route after not found", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/qwerty", endpoint)
		req, err := requests.New(url).Build(t.Context())
		require.NoError(t, err)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
		assert.Empty(t, b)
	})
}
