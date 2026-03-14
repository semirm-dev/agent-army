package runner

// Runner executes shell commands.
type Runner interface {
	Run(cmd string, args ...string) (stdout string, err error)
}
