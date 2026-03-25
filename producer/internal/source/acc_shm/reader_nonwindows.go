//go:build !windows

package acc_shm

import (
	"context"
	"fmt"
)

type Reader interface {
	ReadPhysics(ctx context.Context) (*PhysicsPage, error)
	ReadGraphics(ctx context.Context) (*GraphicsPage, error)
	ReadStatic(ctx context.Context) (*StaticPage, error)
	Close() error
}

type SharedMemoryReader struct{}

func NewReader(ctx context.Context) (*SharedMemoryReader, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("create shared memory reader: %w", err)
	}

	return nil, ErrUnsupportedPlatform
}

func NewSHMReader(ctx context.Context) (*SharedMemoryReader, error) {
	return NewReader(ctx)
}

func (r *SharedMemoryReader) ReadPhysics(ctx context.Context) (*PhysicsPage, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("read physics page: %w", err)
	}

	return nil, ErrUnsupportedPlatform
}

func (r *SharedMemoryReader) ReadGraphics(ctx context.Context) (*GraphicsPage, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("read graphics page: %w", err)
	}

	return nil, ErrUnsupportedPlatform
}

func (r *SharedMemoryReader) ReadStatic(ctx context.Context) (*StaticPage, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("read static page: %w", err)
	}

	return nil, ErrUnsupportedPlatform
}

func (r *SharedMemoryReader) Close() error {
	return nil
}
