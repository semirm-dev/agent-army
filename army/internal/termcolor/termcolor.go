package termcolor

import "fmt"

const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Cyan      = "\033[36m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Red       = "\033[31m"
	Magenta   = "\033[35m"
	BoldCyan  = "\033[1;36m"
	BoldGreen = "\033[1;32m"
)

func Header(label string, count int) string {
	return fmt.Sprintf("\n%s━━━ %s (%d) ━━━%s\n", BoldCyan, label, count, Reset)
}

func Section(name string) string {
	return fmt.Sprintf("  %s[%s]%s", Bold, name, Reset)
}

func Item(name string) string {
	return fmt.Sprintf("    %s•%s %s", Dim, Reset, name)
}

func Numbered(n int, name, extra string) string {
	if extra != "" {
		return fmt.Sprintf("  %s%d.%s %s%s%s %s%s%s", Dim, n, Reset, Bold, name, Reset, Dim, extra, Reset)
	}
	return fmt.Sprintf("  %s%d.%s %s%s%s", Dim, n, Reset, Bold, name, Reset)
}

func Success(msg string) string {
	return fmt.Sprintf("%s✓%s %s", Green, Reset, msg)
}

func Warn(msg string) string {
	return fmt.Sprintf("%s⚠%s %s", Yellow, Reset, msg)
}

func Err(msg string) string {
	return fmt.Sprintf("%s✗%s %s", Red, Reset, msg)
}

func Arrow(msg string) string {
	return fmt.Sprintf("%s→%s %s", Cyan, Reset, msg)
}

func DoneMsg(msg string) string {
	return fmt.Sprintf("\n%s✓ %s%s\n", BoldGreen, msg, Reset)
}

func ErrMsg(msg string) string {
	return fmt.Sprintf("\n%s✗ %s%s\n", Red, msg, Reset)
}
