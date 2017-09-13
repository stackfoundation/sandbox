package properties

import (
	"github.com/magiconair/properties"
)

const placeholderPrefix = "${"
const placeholderSuffix = "}"

// Properties A basic set of properties
type Properties struct {
	m map[string]string
}

// NewProperties Create an empty set of properties
func NewProperties() *Properties {
	return &Properties{
		m: make(map[string]string),
	}
}

// Load Load properties from a file into the properties set
func (p *Properties) Load(file string) error {
	props, err := properties.LoadFile(file, properties.UTF8)
	if err != nil {
		return err
	}

	props.DisableExpansion = true

	for _, key := range props.Keys() {
		p.m[key] = props.GetString(key, "")
	}

	return nil
}

// Map Get properties set as a map (retrieves actual backing map)
func (p *Properties) Map() map[string]string {
	return p.m
}

// Merge Merge properties from another set into this one
func (p *Properties) Merge(other *Properties) {
	if other != nil {
		for k, v := range other.m {
			p.Set(k, v)
		}
	}
}

// ResolveFrom Expands all properties of this set using properties from another
func (p *Properties) ResolveFrom(context *Properties) {
	if context != nil {
		for k, v := range p.m {
			p.m[k] = context.Expand(v)
		}
	}
}

// Set Set a property, expanding the value before setting
func (p *Properties) Set(key string, value string) {
	p.m[key] = p.Expand(value)
}
