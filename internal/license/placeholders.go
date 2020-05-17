package license

import "strings"

// placeholderMap contains a mapping from a field name to known placeholders
// for it in open source license texts.
var placeholderMap = map[string][]string{
	"project": []string{"<program>"},
	"author":  []string{"<name of author>", "[fullname]", "[name of copyright owner]"},
	"year":    []string{"<year>", "[year]", "[yyyy]"},
}

// FieldMap is a map of placeholder field name and replacement values.
type FieldMap map[string]string

// ResolvePlaceholders takes a text and a field map and resolves all
// placeholders to the replacement values in the map.
func ResolvePlaceholders(text string, fieldMap FieldMap) string {
	for fieldName, replacement := range fieldMap {
		placeholders, ok := placeholderMap[fieldName]
		if !ok {
			continue
		}

		for _, placeholder := range placeholders {
			text = strings.ReplaceAll(text, placeholder, replacement)
		}
	}

	return text
}
