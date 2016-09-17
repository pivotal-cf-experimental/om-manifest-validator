package bosh

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type StagedManifestResponse struct {
	Manifest *Manifest `yaml:"manifest"`
}

type Manifest struct {
	Jobs           []*Job           `yaml:"jobs"`
	InstanceGroups []*InstanceGroup `yaml:"instance_groups"`
}

type Job struct {
	N string     `yaml:"name"`
	P Properties `yaml:"properties"`
}

type OMJob interface {
	Name() string
	Properties() Properties
}

func (j *Job) Name() string {
	return j.N
}

func (j *Job) Properties() Properties {
	return j.P
}

func NewJob(name string) *Job {
	return &Job{
		N: name,
	}
}

type InstanceGroup struct {
	N string     `yaml:"name"`
	P Properties `yaml:"properties"`
}

func (ig *InstanceGroup) Name() string {
	return ig.N
}

func (ig *InstanceGroup) Properties() Properties {
	return ig.P
}

func NewInstanceGroup(name string) *InstanceGroup {
	return &InstanceGroup{
		N: name,
	}
}

type Properties map[interface{}]interface{}

func (m *Manifest) JobNamed(name string) (job OMJob) {
	jobName := fmt.Sprintf("%s-partition", name)
	for _, j := range m.Jobs {
		if matched, err := regexp.MatchString("^"+jobName, j.Name()); err == nil && matched {
			job = j
			break
		}
	}
	for _, ig := range m.InstanceGroups {
		if ig.Name() == name {
			job = ig
			break
		}
	}

	if job == nil {
		panic(fmt.Sprintf("Unable to find job named: '%s'", jobName))
	}
	return
}

func (p Properties) Find(lens string) (val interface{}, err error) {
	matchers := strings.Split(lens, ".")

	if len(matchers) == 1 {
		val, found := p[matchers[0]]
		if !found {
			return nil, errors.New("value not found")
		}
		return val, nil
	}

	m := matchers[0]

	if next, present := p[m]; present {
		n, ok := next.(Properties)
		if !ok {
			panic("type conversion failed")
		}
		return n.Find(strings.Join(matchers[1:], "."))
	} else {
		return nil, errors.New("value not found")
	}
}

func (p Properties) FindString(lens string) (val string, err error) {
	s, err := p.Find(lens)
	if err != nil {
		return "", err
	}

	val, ok := s.(string)
	if !ok {
		return "", errors.New("value not a string")
	}

	return val, nil
}

func (p Properties) FindInt(lens string) (val int, err error) {
	s, err := p.Find(lens)
	if err != nil {
		return 0, err
	}

	val, ok := s.(int)
	if !ok {
		return 0, errors.New("value not an integer")
	}

	return val, nil
}

func (p Properties) FindBool(lens string) (val bool, err error) {
	b, err := p.Find(lens)
	if err != nil {
		return false, err
	}

	val, ok := b.(bool)
	if !ok {
		return false, errors.New("value not a boolean")
	}

	return val, nil
}