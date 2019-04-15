package cmds

// ExitStatusAwareError is a custom error type that carries an exit code
type ExitStatusAwareError struct {
	description string
	exitStatus  int
}

// Error returns the error description
func (e *ExitStatusAwareError) Error() string {
	return e.description
}

// ExitStatus returns the exit status
func (e *ExitStatusAwareError) ExitStatus() int {
	return e.exitStatus
}
