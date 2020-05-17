package license

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolvePlaceholders(t *testing.T) {
	text := `This is a made up license text for <program>.

Copyright [year] [fullname]

Copyright [yyyy] [name of copyright owner]

Copyright <year> <name of author>`

	expected := `This is a made up license text for awesome-project.

Copyright 2020 johndoe

Copyright 2020 johndoe

Copyright 2020 johndoe`

	resolved := ResolvePlaceholders(text, FieldMap{
		"project":        "awesome-project",
		"author":         "johndoe",
		"year":           "2020",
		"someotherfield": "someothervalue",
	})

	assert.Equal(t, expected, resolved)
}
