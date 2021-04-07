package kickoff

import (
	"io/ioutil"

	"github.com/ghodss/yaml"
)

// Load loads a file from path into v. Returns an error if reading the file
// fails. Does not perform any defaulting or validation.
func Load(path string, v interface{}) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(buf, &v)
}
