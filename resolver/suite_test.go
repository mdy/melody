package resolver

import (
	"fmt"
	c "gopkg.in/check.v1"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { c.TestingT(t) }

var _ = c.Suite(&MySuite{})

type MySuite struct{}

/* Test Helpers */
func checkSuiteError(err error, args ...interface{}) {
	if err != nil {
		panic(fmt.Sprint(append(args, "\nERROR: ", err)...))
	}
}
