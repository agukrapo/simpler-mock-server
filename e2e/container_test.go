package e2e

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/agukrapo/go-http-client/requests"
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

	t.Run("GET", func(t *testing.T) {
		req, err := requests.New(fmt.Sprintf("http://%s/health", endpoint)).Build(ctx)
		require.NoError(t, err)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, res.StatusCode)
		require.Equal(t, string(b), "simpler-mock-server UP")
	})
}
