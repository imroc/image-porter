package main

import (
	"fmt"
	"regexp"
)

type (
	Config struct {
		Default *Default `yaml:"default,omitempty"`
		Images  []*Image `yaml:"images"`
	}
	Default struct {
		TagFilter *TagFilter `yaml:"tagFilter,omitempty"`
	}
	Image struct {
		TagFilter *TagFilter `yaml:"tagFilter,omitempty"`
		From      string     `yaml:"from"`
		To        string     `yaml:"to"`
	}
	TagFilter struct {
		regex *regexp.Regexp
		Regex string `yaml:"regex"`
	}
)

func (c *Config) Init() (err error) {
	var defaultTagFilter *TagFilter
	if c.Default != nil && c.Default.TagFilter != nil {
		c.Default.TagFilter.Init()
		defaultTagFilter = c.Default.TagFilter
	}
	for _, i := range c.Images {
		if tf := i.TagFilter; tf != nil {
			err = tf.Init()
			if err != nil {
				return
			}
		} else {
			i.TagFilter = defaultTagFilter
		}
	}
	return
}

func (tf *TagFilter) String() string {
	if tf.Regex != "" {
		return fmt.Sprintf(`regex: %s`, tf.Regex)
	}
	return ""
}

func (tf *TagFilter) FilterTags(tags []string) []string {
	if tf == nil {
		return tags
	}
	var result []string
	for _, tag := range tags {
		if tf.Match(tag) {
			result = append(result, tag)
		}
	}
	return result
}

func (tf *TagFilter) Match(tag string) bool {
	if tf.regex != nil {
		return tf.regex.MatchString(tag)
	}
	return false
}

func (tf *TagFilter) Init() (err error) {
	if tf.Regex != "" {
		tf.regex, err = regexp.Compile(tf.Regex)
		return
	}
	return
}
