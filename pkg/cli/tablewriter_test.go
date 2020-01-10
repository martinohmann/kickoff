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

	expected := "ONE       \tTWO        \n00        \t0000000000\t\n1111      \t11111111  \t\n222222    \t222222    \t\n33333333  \t3333      \t\n4444444444\t44        \t\n"

	assert.Equal(t, expected, buf.String())
}
