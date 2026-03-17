package shared_memory

// Shared Memory Names - Windows Named Pipes
const (
	PhysicsMemoryName = "Local\\acpmf_physics"
	StaticMemoryName  = "Local\\acpmf_static"
	GraphicsMemoryName = "Local\\acpmf_graphics"
)

// Data Sizes in bytes
const (
	PhysicsDataSize   = 740  // SPageFilePhysics
	StaticDataSize    = 256  // SPageFileStatic
	GraphicsDataSize  = 512  // SPageFileGraphic
)