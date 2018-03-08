package configs

// Config format
type Config struct {
	ContainerEvents  bool
	ProcessEvents    bool
	NetworkEvents    bool
	SyscallEvents    bool
	KernelCallEvents bool
	FileEvents       bool
	FilePatterns     []string
}
