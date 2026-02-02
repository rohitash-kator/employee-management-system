package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	client *mongo.Client
}

func Connect(ctx context.Context, uri string, timeout time.Duration) (*Client, error) {
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cl, err := mongo.Connect(cctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}
	if err := cl.Ping(cctx, nil); err != nil {
		_ = cl.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	return &Client{client: cl}, nil
}

func (c *Client) Database(name string) *mongo.Database {
	return c.client.Database(name)
}

func (c *Client) Disconnect(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}
