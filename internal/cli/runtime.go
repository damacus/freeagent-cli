package cli

type Runtime struct {
	ConfigPath      string
	Profile         string
	Sandbox         bool
	BaseURL         string
	BaseURLOverride bool
	JSONOutput      bool
}
