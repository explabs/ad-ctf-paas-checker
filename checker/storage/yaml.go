package storage

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type ServicesInfo struct {
	Services []*Service `yaml:"checker"`
}

func (s *ServicesInfo) Load() error {
	buf, err := ioutil.ReadFile("config.yml")
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(buf, &s)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}
