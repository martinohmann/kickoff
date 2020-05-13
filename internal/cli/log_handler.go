package cli

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/apex/log"
	"github.com/fatih/color"
	colorable "github.com/mattn/go-colorable"
)

var bold = color.New(color.Bold)

// Colors mapping.
var Colors = [...]*color.Color{
	log.DebugLevel: color.New(color.FgGreen),
	log.InfoLevel:  color.New(color.FgBlue),
	log.WarnLevel:  color.New(color.FgYellow),
	log.ErrorLevel: color.New(color.FgRed),
	log.FatalLevel: color.New(color.FgRed),
}

// Strings mapping.
var Strings = [...]string{
	log.DebugLevel: "•",
	log.InfoLevel:  "•",
	log.WarnLevel:  "•",
	log.ErrorLevel: "⨯",
	log.FatalLevel: "⨯",
}

// LogHandler implementation.
type LogHandler struct {
	mu      sync.Mutex
	Writer  io.Writer
	Padding int
}

// NewLogHandler creates a new LogHandler which writes to w.
func NewLogHandler(w io.Writer) *LogHandler {
	if f, ok := w.(*os.File); ok {
		return &LogHandler{
			Writer:  colorable.NewColorable(f),
			Padding: 0,
		}
	}

	return &LogHandler{
		Writer:  w,
		Padding: 0,
	}
}

// HandleLog implements log.Handler.
func (h *LogHandler) HandleLog(e *log.Entry) error {
	color := Colors[e.Level]
	level := Strings[e.Level]
	names := e.Fields.Names()

	h.mu.Lock()
	defer h.mu.Unlock()

	color.Fprintf(h.Writer, "%s %-32s", bold.Sprintf("%*s", h.Padding+1, level), e.Message)

	for _, name := range names {
		if name == "source" {
			continue
		}

		fmt.Fprintf(h.Writer, "  %s=%v", color.Sprint(name), e.Fields.Get(name))
	}

	fmt.Fprintln(h.Writer)

	return nil
}
