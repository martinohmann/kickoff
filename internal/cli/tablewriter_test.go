package cli

import (
	"bytes"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTableWriter_Render(t *testing.T) {
	var buf bytes.Buffer

	tw := NewTableWriter(&buf)
	tw.SetHeader("One", "Two")

	for i := 0; i < 5; i++ {
		one := strings.Repeat(strconv.Itoa(i), (i+1)*2)
		two := strings.Repeat(strconv.Itoa(i), (5-i)*2)

		tw.Append(one, two)
	}

	tw.Render()

	expected := "One        Two        \n00         0000000000 \n1111       11111111   \n222222     222222     \n33333333   3333       \n4444444444 44         \n"

	assert.Equal(t, expected, buf.String())
}
