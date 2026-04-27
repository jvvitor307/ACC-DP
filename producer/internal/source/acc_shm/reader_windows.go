//go:build windows

package acc_shm

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type Reader interface {
	ReadPhysics(ctx context.Context) (*PhysicsRawPage, error)
	ReadGraphics(ctx context.Context) (*GraphicsRawPage, error)
	ReadStatic(ctx context.Context) (*StaticRawPage, error)
	Close() error
}

type SharedMemoryReader struct {
	mu       sync.RWMutex
	closed   bool
	physics  mappedPage
	graphics mappedPage
	static   mappedPage
}

type mappedPage struct {
	name   string
	handle windows.Handle
	addr   uintptr
	size   int
}

var (
	kernel32DLL          = windows.NewLazySystemDLL("kernel32.dll")
	procOpenFileMappingW = kernel32DLL.NewProc("OpenFileMappingW")
)

func NewReader(ctx context.Context) (*SharedMemoryReader, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("create shared memory reader: %w", err)
	}

	if err := validateTypeLayout(); err != nil {
		return nil, fmt.Errorf("create shared memory reader: %w", err)
	}

	reader := &SharedMemoryReader{}

	physics, err := openMappedPage(PhysicsSharedMemoryName, PhysicsPageSize)
	if err != nil {
		return nil, fmt.Errorf("open physics shared memory: %w", err)
	}
	reader.physics = physics

	graphics, err := openMappedPage(GraphicsSharedMemoryName, GraphicsPageSize)
	if err != nil {
		_ = reader.Close()
		return nil, fmt.Errorf("open graphics shared memory: %w", err)
	}
	reader.graphics = graphics

	staticPage, err := openMappedPage(StaticSharedMemoryName, StaticPageSize)
	if err != nil {
		_ = reader.Close()
		return nil, fmt.Errorf("open static shared memory: %w", err)
	}
	reader.static = staticPage

	return reader, nil
}

func NewSHMReader(ctx context.Context) (*SharedMemoryReader, error) {
	return NewReader(ctx)
}

func (r *SharedMemoryReader) ReadPhysics(ctx context.Context) (*PhysicsRawPage, error) {
	data, err := r.readPage(ctx, &r.physics)
	if err != nil {
		return nil, fmt.Errorf("read physics page: %w", err)
	}

	page, err := DecodePhysicsPage(data)
	if err != nil {
		return nil, fmt.Errorf("read physics page: %w", err)
	}

	return page, nil
}

func (r *SharedMemoryReader) ReadGraphics(ctx context.Context) (*GraphicsRawPage, error) {
	data, err := r.readPage(ctx, &r.graphics)
	if err != nil {
		return nil, fmt.Errorf("read graphics page: %w", err)
	}

	page, err := DecodeGraphicsPage(data)
	if err != nil {
		return nil, fmt.Errorf("read graphics page: %w", err)
	}

	return page, nil
}

func (r *SharedMemoryReader) ReadStatic(ctx context.Context) (*StaticRawPage, error) {
	data, err := r.readPage(ctx, &r.static)
	if err != nil {
		return nil, fmt.Errorf("read static page: %w", err)
	}

	page, err := DecodeStaticPage(data)
	if err != nil {
		return nil, fmt.Errorf("read static page: %w", err)
	}

	return page, nil
}

func (r *SharedMemoryReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	var closeErr error
	closeErr = errors.Join(closeErr, closeMappedPage(&r.physics))
	closeErr = errors.Join(closeErr, closeMappedPage(&r.graphics))
	closeErr = errors.Join(closeErr, closeMappedPage(&r.static))

	r.closed = true
	return closeErr
}

func (r *SharedMemoryReader) readPage(ctx context.Context, page *mappedPage) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return nil, ErrReaderClosed
	}

	if page.addr == 0 || page.size <= 0 {
		return nil, fmt.Errorf("mapped page is not available")
	}

	raw := unsafe.Slice((*byte)(unsafe.Pointer(page.addr)), page.size)
	data := make([]byte, page.size)
	copy(data, raw)

	return data, nil
}

func openMappedPage(name string, size int) (mappedPage, error) {
	namePtr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return mappedPage{}, fmt.Errorf("convert shared memory name: %w", err)
	}

	handle, err := openFileMapping(windows.FILE_MAP_READ, false, namePtr)
	if err != nil {
		return mappedPage{}, err
	}

	addr, err := windows.MapViewOfFile(handle, windows.FILE_MAP_READ, 0, 0, uintptr(size))
	if err != nil {
		_ = windows.CloseHandle(handle)
		return mappedPage{}, err
	}

	return mappedPage{
		name:   name,
		handle: handle,
		addr:   addr,
		size:   size,
	}, nil
}

func openFileMapping(access uint32, inheritHandle bool, name *uint16) (windows.Handle, error) {
	var inherit uintptr
	if inheritHandle {
		inherit = 1
	}

	r1, _, callErr := procOpenFileMappingW.Call(
		uintptr(access),
		inherit,
		uintptr(unsafe.Pointer(name)),
	)
	if r1 == 0 {
		if callErr != nil && callErr != windows.ERROR_SUCCESS {
			return 0, callErr
		}
		return 0, syscall.EINVAL
	}

	return windows.Handle(r1), nil
}

func closeMappedPage(page *mappedPage) error {
	if page == nil {
		return nil
	}

	var closeErr error
	if page.addr != 0 {
		if err := windows.UnmapViewOfFile(page.addr); err != nil {
			closeErr = errors.Join(closeErr, fmt.Errorf("unmap %s: %w", page.name, err))
		}
		page.addr = 0
	}

	if page.handle != 0 {
		if err := windows.CloseHandle(page.handle); err != nil {
			closeErr = errors.Join(closeErr, fmt.Errorf("close handle %s: %w", page.name, err))
		}
		page.handle = 0
	}

	return closeErr
}
