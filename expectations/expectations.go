package expectations

import (
	"fmt"
)

type Expectation struct {
	actual interface{}
}

func Expect(actual interface{}) *Expectation {
	return &Expectation{actual: actual}
}

func (e *Expectation) ToEqual(expected interface{}) (bool, string) {
	if e.actual == expected {
		return true, fmt.Sprintf("Got: %v", e.actual)
	}
	return false, fmt.Sprintf("Expected %v to equal %v but it doesn't", expected, e.actual)
}

func (e *Expectation) ToBeGreaterThan(expected interface{}) (bool, string) {
	switch actual := e.actual.(type) {
	case int:
		if exp, ok := expected.(int); ok {
			if actual > exp {
				return true, fmt.Sprintf("Got: %v (> %v)", actual, exp)
			}
		}
	case float64:
		if exp, ok := expected.(float64); ok {
			if actual > exp {
				return true, fmt.Sprintf("Got: %v (> %v)", actual, exp)
			}
		}
	}
	return false, fmt.Sprintf("Expected %v to be greater than %v but it's not", e.actual, expected)
}

func (e *Expectation) ToBeLessThan(expected interface{}) (bool, string) {
	switch actual := e.actual.(type) {
	case int:
		if exp, ok := expected.(int); ok {
			if actual < exp {
				return true, fmt.Sprintf("Got: %v (< %v)", actual, exp)
			}
		}
	case float64:
		if exp, ok := expected.(float64); ok {
			if actual < exp {
				return true, fmt.Sprintf("Got: %v (< %v)", actual, exp)
			}
		}
	}
	return false, fmt.Sprintf("Expected %v to be less than %v but it's not", e.actual, expected)
}

func MakeSure(condition bool, message string) (bool, string) {
	if condition {
		return true, message
	}
	return false, fmt.Sprintf("Check failed: %s", message)
}
