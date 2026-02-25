//go:build integration
// +build integration

package app_test

import (
	"testing"

	app "github.com/EfosaE/credora-backend/di"
	"github.com/EfosaE/credora-backend/test"
	"github.com/stretchr/testify/require"
)

func TestAppBuilder_BuildVariants(t *testing.T) {
	cfg := test.NewTestConfig()

	dbPool, dbCleanup := test.SetupTestDB(t)
	defer dbCleanup()

	redisClient, redisCleanup := test.SetupTestRedis(t)
	defer redisCleanup()

	tests := []struct {
		name    string
		buildFn func(*app.AppBuilder) (*app.AppDependencies, error)
	}{
		{
			name: "BuildForServer",
			buildFn: func(b *app.AppBuilder) (*app.AppDependencies, error) {
				return b.
					WithLogger("server").
					WithDBFromPool(dbPool).
					WithRedisFromClient(redisClient).
					WithEventBus().
					WithQueueClient().
					WithRepositories().
					Build()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := app.NewAppBuilder(cfg)

			deps, err := tt.buildFn(builder)

			require.NoError(t, err)
			require.NotNil(t, deps)
		})
	}
}
