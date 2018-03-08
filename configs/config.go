package configs

// Config format
type Config struct {
	ContainerEvents  bool
	SystemdEvents    bool
	CacheMissEvents  bool
	ProcessEvents    bool
	NetworkEvents    bool
	SyscallEvents    bool
	KernelCallEvents bool
	FileEvents       bool
	FilePatterns     []string
}
