package acc_shm

import (
	"errors"
	"fmt"
	"unsafe"
)

func DecodePhysicsPage(data []byte) (*PhysicsPage, error) {
	page, err := decodePage[PhysicsPage](data, PhysicsPageSize, ErrInvalidPhysicsSize)
	if err != nil {
		return nil, fmt.Errorf("decode physics page: %w", err)
	}

	return page, nil
}

func DecodeGraphicsPage(data []byte) (*GraphicsPage, error) {
	page, err := decodePage[GraphicsPage](data, GraphicsPageSize, ErrInvalidGraphicsSize)
	if err != nil {
		return nil, fmt.Errorf("decode graphics page: %w", err)
	}

	return page, nil
}

func DecodeStaticPage(data []byte) (*StaticPage, error) {
	page, err := decodePage[StaticPage](data, StaticPageSize, ErrInvalidStaticSize)
	if err != nil {
		return nil, fmt.Errorf("decode static page: %w", err)
	}

	return page, nil
}

func decodePage[T any](data []byte, expectedSize int, sizeErr error) (*T, error) {
	if len(data) != expectedSize {
		return nil, fmt.Errorf("%w: got=%d want=%d", sizeErr, len(data), expectedSize)
	}

	var page T
	if int(unsafe.Sizeof(page)) != expectedSize {
		return nil, fmt.Errorf("layout size mismatch: got=%d want=%d", unsafe.Sizeof(page), expectedSize)
	}

	dst := unsafe.Slice((*byte)(unsafe.Pointer(&page)), expectedSize)
	copy(dst, data)

	return &page, nil
}

func validateTypeLayout() error {
	physicsSize := int(unsafe.Sizeof(PhysicsPage{}))
	if physicsSize != PhysicsPageSize {
		return fmt.Errorf("physics page layout mismatch: got=%d want=%d", physicsSize, PhysicsPageSize)
	}

	graphicsSize := int(unsafe.Sizeof(GraphicsPage{}))
	if graphicsSize != GraphicsPageSize {
		return fmt.Errorf("graphics page layout mismatch: got=%d want=%d", graphicsSize, GraphicsPageSize)
	}

	staticSize := int(unsafe.Sizeof(StaticPage{}))
	if staticSize != StaticPageSize {
		return fmt.Errorf("static page layout mismatch: got=%d want=%d", staticSize, StaticPageSize)
	}

	return nil
}

func isSizeError(err error) bool {
	return errors.Is(err, ErrInvalidPhysicsSize) ||
		errors.Is(err, ErrInvalidGraphicsSize) ||
		errors.Is(err, ErrInvalidStaticSize)
}
