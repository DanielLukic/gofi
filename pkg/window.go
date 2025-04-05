package pkg

// Window represents a window in the window manager
type Window struct {
	ID      string
	Desktop int
	PID     string
	Command string
	Class   string // Short window class (e.g. thunderbird)
	Name    string // Full window class with instance (e.g. Mail.thunderbird)
	Title   string
}
