package license

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	licensesURL = "https://api.github.com/licenses"
)

var (
	ErrLicenseNotFound = errors.New("license not found")
)

type Info struct {
	Key  string
	Name string
	Body string
}

func Lookup(name string) (*Info, error) {
	name = strings.ToLower(name)

	licenses, err := List()
	if err != nil {
		return nil, err
	}

	for _, license := range licenses {
		if strings.ToLower(license.Key) == name || strings.ToLower(license.Name) == name {
			return Get(license.Key)
		}
	}

	return nil, ErrLicenseNotFound
}

func Get(key string) (*Info, error) {
	info := Info{}

	err := get(fmt.Sprintf("%s/%s", licensesURL, key), &info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func List() ([]*Info, error) {
	infos := []*Info{}

	err := get(licensesURL, &infos)
	if err != nil {
		return nil, err
	}

	return infos, nil
}

func get(url string, into interface{}) error {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	r.Header.Add("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, into)
}
