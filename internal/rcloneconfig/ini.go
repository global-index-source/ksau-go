package rcloneconfig

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

type Config struct {
	Sections map[string]*Section
	Order    []string
}

type Section struct {
	Name     string
	Keys     map[string]string
	KeyOrder []string
}

func NewConfig() *Config {
	return &Config{Sections: make(map[string]*Section)}
}

func Parse(data []byte) (*Config, error) {
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(content, "\n")
	config := NewConfig()
	var current *Section

	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			sectionName := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]"))
			if sectionName == "" {
				return nil, fmt.Errorf("empty section name at line %d", idx+1)
			}
			current = config.EnsureSection(sectionName)
			continue
		}

		if current == nil {
			return nil, fmt.Errorf("key-value outside section at line %d", idx+1)
		}

		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) == 1 {
			key := strings.TrimSpace(parts[0])
			if key == "" {
				return nil, fmt.Errorf("empty key at line %d", idx+1)
			}
			current.Set(key, "")
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("empty key at line %d", idx+1)
		}
		current.Set(key, value)
	}

	return config, nil
}

func (c *Config) EnsureSection(name string) *Section {
	if section, ok := c.Sections[name]; ok {
		return section
	}
	section := &Section{Name: name, Keys: make(map[string]string)}
	c.Sections[name] = section
	c.Order = append(c.Order, name)
	return section
}

func (c *Config) RemoveSection(name string) bool {
	if _, ok := c.Sections[name]; !ok {
		return false
	}
	delete(c.Sections, name)
	for i, entry := range c.Order {
		if entry == name {
			c.Order = append(c.Order[:i], c.Order[i+1:]...)
			break
		}
	}
	return true
}

func (c *Config) Render() []byte {
	var buffer bytes.Buffer
	for _, name := range c.Order {
		section, ok := c.Sections[name]
		if !ok {
			continue
		}
		buffer.WriteString("[")
		buffer.WriteString(section.Name)
		buffer.WriteString("]\n")
		keys := section.keyOrderWithFallback()
		for _, key := range keys {
			value := section.Keys[key]
			buffer.WriteString(key)
			buffer.WriteString(" = ")
			buffer.WriteString(value)
			buffer.WriteString("\n")
		}
		buffer.WriteString("\n")
	}
	return buffer.Bytes()
}

func (s *Section) Set(key, value string) {
	if _, ok := s.Keys[key]; !ok {
		s.KeyOrder = append(s.KeyOrder, key)
	}
	s.Keys[key] = value
}

func (s *Section) keyOrderWithFallback() []string {
	if len(s.KeyOrder) == 0 {
		keys := make([]string, 0, len(s.Keys))
		for key := range s.Keys {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		return keys
	}
	return s.KeyOrder
}
