package codec

import "fmt"

type parseError struct {
	line int
	msg  string
}

func (e parseError) Error() string {
	return fmt.Sprintf("line %d: %s", e.line, e.msg)
}

func errorAt(line int, msg string) error {
	return parseError{line: line, msg: msg}
}

func errorAtf(line int, format string, args ...any) error {
	return parseError{line: line, msg: fmt.Sprintf(format, args...)}
}

func errorWrap(line int, err error) error {
	if err == nil {
		return nil
	}
	return parseError{line: line, msg: err.Error()}
}
