package provider

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func newRedisOpts(env *EnvProvider) redis.Options {
	redisURL := env.redisHost + ":" + env.redisPort

	opts := redis.Options{
		Addr:     redisURL,
		Password: env.redisPassword,
		DB:       env.redisDB,
	}

	if (env.appEnv != "local") && (env.appEnv != "test") {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	return opts
}

func NewRedisProvider(env *EnvProvider) (*redis.Client, error) {
	opts := newRedisOpts(env)
	rdb := redis.NewClient(&opts)

	err := rdb.Ping(context.Background()).Err()
	if err != nil {
		return nil, fmt.Errorf(
			"unable to initialize redis connection (addr: %s): %w",
			opts.Addr,
			err,
		)
	}

	return rdb, nil
}
