package shared_memory

// Shared Memory Names - Windows Named Pipes
const (
	PhysicsMemoryName = "Local\\acpmf_physics"
	StaticMemoryName  = "Local\\acpmf_static"
	GraphicsMemoryName = "Local\\acpmf_graphics"
)

// Data Sizes in bytes
const (
	PhysicsDataSize   = 800
	StaticDataSize    = 784
	GraphicsDataSize  = 1588
)