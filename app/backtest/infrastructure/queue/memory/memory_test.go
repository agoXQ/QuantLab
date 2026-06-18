package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	domqueue "github.com/agoXQ/QuantLab/app/backtest/domain/queue"
)

func TestEnqueueDequeue(t *testing.T) {
	q := New(2)
	defer q.Close()
	ctx := context.Background()

	if err := q.Enqueue(ctx, domqueue.Job{ID: 7}); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	got, err := q.Dequeue(ctx)
	if err != nil {
		t.Fatalf("Dequeue: %v", err)
	}
	if got.ID != 7 {
		t.Fatalf("expected job 7, got %+v", got)
	}
}

func TestDequeueRespectsContext(t *testing.T) {
	q := New(1)
	defer q.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := q.Dequeue(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestEnqueueAfterCloseFails(t *testing.T) {
	q := New(1)
	if err := q.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := q.Enqueue(context.Background(), domqueue.Job{ID: 1}); !errors.Is(err, domqueue.ErrQueueClosed) {
		t.Fatalf("expected ErrQueueClosed, got %v", err)
	}
}

func TestDequeueAfterCloseUnblocks(t *testing.T) {
	q := New(1)
	done := make(chan error, 1)
	go func() {
		_, err := q.Dequeue(context.Background())
		done <- err
	}()
	time.Sleep(20 * time.Millisecond)
	_ = q.Close()
	select {
	case err := <-done:
		if !errors.Is(err, domqueue.ErrQueueClosed) {
			t.Fatalf("expected ErrQueueClosed, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Dequeue did not unblock after Close")
	}
}
