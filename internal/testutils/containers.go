package testutils

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestContainers содержит все контейнеры для тестирования
type TestContainers struct {
	Postgres testcontainers.Container
	Redis    testcontainers.Container
	MongoDB  testcontainers.Container
	RabbitMQ testcontainers.Container
}

// StartTestContainers запускает все необходимые контейнеры для тестирования
func StartTestContainers(ctx context.Context) (*TestContainers, error) {
	containers := &TestContainers{}

	// Запуск PostgreSQL
	postgres, err := startPostgres(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres: %w", err)
	}
	containers.Postgres = postgres

	// Запуск Redis
	redis, err := startRedis(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start redis: %w", err)
	}
	containers.Redis = redis

	// Запуск MongoDB
	mongodb, err := startMongoDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start mongodb: %w", err)
	}
	containers.MongoDB = mongodb

	// Запуск RabbitMQ
	rabbitmq, err := startRabbitMQ(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start rabbitmq: %w", err)
	}
	containers.RabbitMQ = rabbitmq

	return containers, nil
}

// StopTestContainers останавливает все контейнеры
func (tc *TestContainers) StopTestContainers(ctx context.Context) error {
	var errors []error

	if tc.Postgres != nil {
		if err := tc.Postgres.Terminate(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop postgres: %w", err))
		}
	}

	if tc.Redis != nil {
		if err := tc.Redis.Terminate(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop redis: %w", err))
		}
	}

	if tc.MongoDB != nil {
		if err := tc.MongoDB.Terminate(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop mongodb: %w", err))
		}
	}

	if tc.RabbitMQ != nil {
		if err := tc.RabbitMQ.Terminate(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop rabbitmq: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors stopping containers: %v", errors)
	}

	return nil
}

// startPostgres запускает PostgreSQL контейнер
func startPostgres(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "lootbay_test",
			"POSTGRES_USER":     "lootbay",
			"POSTGRES_PASSWORD": "password",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, err
	}

	log.Printf("PostgreSQL started at %s:%s", host, port.Port())
	return container, nil
}

// startRedis запускает Redis контейнер
func startRedis(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		return nil, err
	}

	log.Printf("Redis started at %s:%s", host, port.Port())
	return container, nil
}

// startMongoDB запускает MongoDB контейнер
func startMongoDB(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "mongo:6",
		ExposedPorts: []string{"27017/tcp"},
		Env: map[string]string{
			"MONGO_INITDB_ROOT_USERNAME": "lootbay",
			"MONGO_INITDB_ROOT_PASSWORD": "password",
			"MONGO_INITDB_DATABASE":      "lootbay_chat_test",
		},
		WaitingFor: wait.ForLog("Waiting for connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := container.MappedPort(ctx, "27017")
	if err != nil {
		return nil, err
	}

	log.Printf("MongoDB started at %s:%s", host, port.Port())
	return container, nil
}

// startRabbitMQ запускает RabbitMQ контейнер
func startRabbitMQ(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3-management-alpine",
		ExposedPorts: []string{"5672/tcp", "15672/tcp"},
		Env: map[string]string{
			"RABBITMQ_DEFAULT_USER": "lootbay",
			"RABBITMQ_DEFAULT_PASS": "password",
		},
		WaitingFor: wait.ForLog("Server startup complete"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := container.MappedPort(ctx, "5672")
	if err != nil {
		return nil, err
	}

	log.Printf("RabbitMQ started at %s:%s", host, port.Port())
	return container, nil
}

// GetContainerConnectionInfo возвращает информацию о подключении к контейнерам
func (tc *TestContainers) GetContainerConnectionInfo(ctx context.Context) (map[string]string, error) {
	info := make(map[string]string)

	// PostgreSQL
	if tc.Postgres != nil {
		host, err := tc.Postgres.Host(ctx)
		if err != nil {
			return nil, err
		}
		port, err := tc.Postgres.MappedPort(ctx, "5432")
		if err != nil {
			return nil, err
		}
		info["POSTGRES_HOST"] = host
		info["POSTGRES_PORT"] = port.Port()
	}

	// Redis
	if tc.Redis != nil {
		host, err := tc.Redis.Host(ctx)
		if err != nil {
			return nil, err
		}
		port, err := tc.Redis.MappedPort(ctx, "6379")
		if err != nil {
			return nil, err
		}
		info["REDIS_HOST"] = host
		info["REDIS_PORT"] = port.Port()
	}

	// MongoDB
	if tc.MongoDB != nil {
		host, err := tc.MongoDB.Host(ctx)
		if err != nil {
			return nil, err
		}
		port, err := tc.MongoDB.MappedPort(ctx, "27017")
		if err != nil {
			return nil, err
		}
		info["MONGODB_HOST"] = host
		info["MONGODB_PORT"] = port.Port()
	}

	// RabbitMQ
	if tc.RabbitMQ != nil {
		host, err := tc.RabbitMQ.Host(ctx)
		if err != nil {
			return nil, err
		}
		port, err := tc.RabbitMQ.MappedPort(ctx, "5672")
		if err != nil {
			return nil, err
		}
		info["RABBITMQ_HOST"] = host
		info["RABBITMQ_PORT"] = port.Port()
	}

	return info, nil
}
