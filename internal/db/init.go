package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/Fekinox/chrysalis-backend/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func InitTestDB(
	config *config.Config,
) (*dockertest.Pool, *dockertest.Resource, error) {
	fmt.Println("initializing test db")
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, err
	}

	err = pool.Client.Ping()
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("successfully pinged pool")

	testDBResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "alpine",
		Env: []string{
			"POSTGRES_PASSWORD=" + config.DBPassword,
			"POSTGRES_USER=" + config.DBUsername,
			"POSTGRES_PASSWORD=" + config.DBPassword,
			"POSTGRES_DB=" + config.DBName,
			"listen_addresses = '*'",
		},
	}, func(hc *docker.HostConfig) {
		hc.AutoRemove = true
		hc.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})

	if err != nil {
		return nil, nil, err
	}
	fmt.Println("successfully created db resource")

	if err := pool.Retry(func() error {
		fmt.Println("pinging...")
		var err error
		conn, err := pgx.Connect(
			context.Background(),
			fmt.Sprintf(
				"postgres://%s:%s@localhost:%s/%s",
				config.DBUsername,
				config.DBPassword,
				testDBResource.GetPort("5432/tcp"),
				config.DBName,
			),
		)
		if err != nil {
			return err
		}

		return conn.Ping(context.Background())
	}); err != nil {
		return nil, nil, err
	}

	hostAndPort := testDBResource.GetHostPort("5432/tcp")
	newHostAndPort := strings.Split(hostAndPort, ":")
	config.DBHost = newHostAndPort[0]
	config.DBPort = newHostAndPort[1]

	return pool, testDBResource, nil
}

func AutoMigrate(cfg *config.Config) error {
	migrationPath := fmt.Sprintf("file://%s", cfg.MigrationPath)

	m, err := migrate.New(migrationPath, cfg.GetDBUrl())
	if err != nil {
		return err
	}

	if cfg.Recreate {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			return err
		}
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return err
}
