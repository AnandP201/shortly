package generator

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"time"
)

const (
	SeqBits         = 36
	MaxSeq          = (uint64(1) << SeqBits) - 1 // 1,099,511,627,775
	CheckpointBatch = uint64(100_000)            // Async checkpoint every 100k requests
)

type InMemoryCounter struct {
	current    uint64
	regionID   uint64
	workerID   uint64
	store      StateStore
	lastSaved  uint64
	syncSignal chan uint64
	stopChan   chan struct{}
}

func NewInMemoryCounter(regionID, workerID, startSeq uint64, store StateStore) *InMemoryCounter {
	c := &InMemoryCounter{
		current:    startSeq,
		regionID:   regionID,
		workerID:   workerID,
		store:      store,
		lastSaved:  startSeq,
		syncSignal: make(chan uint64, 100),
		stopChan:   make(chan struct{}),
	}

	go c.startAsyncCheckpointer()
	return c
}

// Next increments the sequence atomically with zero thread locks.
func (c *InMemoryCounter) Next() (uint64, error) {
	seq := atomic.AddUint64(&c.current, 1)
	if seq > MaxSeq {
		return 0, errors.New("40-bit sequence capacity exhausted for this pod")
	}

	// Trigger async checkpoint at batch boundaries without blocking caller
	if seq%CheckpointBatch == 0 {
		select {
		case c.syncSignal <- seq:
		default:
			// Non-blocking drop: another sync is already queued
		}
	}

	return seq, nil
}

func (c *InMemoryCounter) startAsyncCheckpointer() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			// Final flush on graceful shutdown
			curr := atomic.LoadUint64(&c.current)
			_ = c.store.SaveCheckpoint(context.Background(), c.regionID, c.workerID, curr)
			return

		case seq := <-c.syncSignal:
			c.persist(seq)

		case <-ticker.C:
			curr := atomic.LoadUint64(&c.current)
			if curr > c.lastSaved {
				c.persist(curr)
			}
		}
	}
}

func (c *InMemoryCounter) persist(seq uint64) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := c.store.SaveCheckpoint(ctx, c.regionID, c.workerID, seq); err != nil {
		log.Printf("Warning: failed to persist checkpoint at %d: %v", seq, err)
	} else {
		c.lastSaved = seq
	}
}

func (c *InMemoryCounter) Close() {
	close(c.stopChan)
}
