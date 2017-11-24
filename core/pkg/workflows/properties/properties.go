package properties

import (
	"github.com/magiconair/properties"
	"github.com/stackfoundation/core/pkg/workflows/errors"
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
func (p *Properties) Merge(other *Properties) error {
	if other != nil {
		composite := errors.NewCompositeError()
		for k, v := range other.m {
			composite.Append(p.Set(k, v))
		}

		return composite.OrNilIfEmpty()
	}

	return nil
}

// ResolveFrom Expands all properties of this set using properties from another
func (p *Properties) ResolveFrom(context *Properties) error {
	if context != nil {
		composite := errors.NewCompositeError()
		for k, v := range p.m {
			v, err := context.Expand(v)
			composite.Append(err)

			p.m[k] = v
		}

		return composite.OrNilIfEmpty()
	}

	return nil
}

// Set Set a property, expanding the value before setting
func (p *Properties) Set(key string, value string) error {
	value, err := p.Expand(value)
	p.m[key] = value
	return err
}
