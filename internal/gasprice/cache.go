package gasprice

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type GasPricer interface {
	GasPrice(ctx context.Context) (float64, float64, error)
}

type CacheGasPricer struct {
	ttl     time.Duration
	backend GasPricer

	mu              sync.Mutex
	expireAt        time.Time
	maxGasPriceGwei float64
	tipCapGwei      float64
}

func NewCacheGasPricer(backend GasPricer, ttl time.Duration) *CacheGasPricer {
	return &CacheGasPricer{
		ttl:     ttl,
		backend: backend,
	}
}

func (c *CacheGasPricer) GasPrice(ctx context.Context) (float64, float64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.expireAt.Before(time.Now()) {
		err := c.updateGasPrice(ctx)
		if err != nil {
			return 0, 0, fmt.Errorf("update gas price: %w", err)
		}
	}

	return c.maxGasPriceGwei, c.tipCapGwei, nil
}

func (c *CacheGasPricer) updateGasPrice(ctx context.Context) error {
	maxGasPriceGwei, tipCapGwei, err := c.backend.GasPrice(ctx)
	if err != nil {
		return err
	}

	c.expireAt = time.Now().Add(c.ttl)
	c.maxGasPriceGwei = maxGasPriceGwei
	c.tipCapGwei = tipCapGwei
	return nil
}
