//go:build integration

package test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func SetupTestRedis(t *testing.T) (*redis.Client, func()) {
	ctx := context.Background()

	container, err := tcredis.Run(ctx, "redis:7-alpine")
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Printf("failed to terminate redis container: %s", err)
		}
	})

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "6379")
	require.NoError(t, err)

	addr := fmt.Sprintf("%s:%s", host, port.Port())

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	require.NoError(t, client.Ping(ctx).Err())

	return client, func() {
		if err := client.Close(); err != nil {
			log.Printf("failed to close redis client: %s", err)
		}
	}
}
