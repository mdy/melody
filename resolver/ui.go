package resolver

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"
)

type UI interface {
	ProgressRate() time.Duration
	IndicateProgress()
	BeforeResolution()
	AfterResolution()
	Debug(int, ...interface{})
}

func NewStdoutUI() UI {
	return &writerUI{os.Stdout, false}
}

type writerUI struct {
	io.Writer
	debug bool
}

func (ui *writerUI) ProgressRate() time.Duration {
	return 333 * time.Millisecond
}

func (ui *writerUI) IndicateProgress() {
	fmt.Fprintf(ui, ".")
}

func (ui *writerUI) BeforeResolution() {
	fmt.Fprintf(ui, "♫ Resolving dependencies...")
}

func (ui *writerUI) AfterResolution() {
	fmt.Fprintln(ui, " done.")
}

func (ui *writerUI) Debug(depth int, args ...interface{}) {
	if !ui.debug || len(args) == 0 {
		return
	}
	prefix := strings.Repeat(" ", depth)
	output, isString := args[0].(string)

	if isString {
		output = fmt.Sprintf(output, args[1:]...)
	} else {
		output = fmt.Sprintf("%s %+v", reflect.TypeOf(args[0]), args[0])
	}

	for _, line := range strings.Split(output, "\n") {
		fmt.Fprintln(ui, prefix+line)
	}
}
