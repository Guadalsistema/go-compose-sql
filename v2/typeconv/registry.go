package typeconv

import (
	"database/sql"
	"fmt"
	"reflect"
)

// ConverterFunc is a function that converts a source value to a target type
type ConverterFunc func(source interface{}) (interface{}, error)

// TypePair represents a pair of source and target types
type TypePair struct {
	Source reflect.Type
	Target reflect.Type
}

// Registry manages type conversions between database types and Go types
type Registry struct {
	converters map[TypePair]ConverterFunc
	defaults   map[reflect.Type]ConverterFunc
}

// NewRegistry creates a new type converter registry
func NewRegistry() *Registry {
	return &Registry{
		converters: make(map[TypePair]ConverterFunc),
		defaults:   make(map[reflect.Type]ConverterFunc),
	}
}

// Register registers a specific converter for a source->target type pair
func (r *Registry) Register(sourceType, targetType reflect.Type, converter ConverterFunc) {
	r.converters[TypePair{Source: sourceType, Target: targetType}] = converter
}

// RegisterDefault registers a default converter for a target type
// This converter will be used when no specific converter is found
func (r *Registry) RegisterDefault(targetType reflect.Type, converter ConverterFunc) {
	r.defaults[targetType] = converter
}

// Convert converts a source value to the target type using registered converters
func (r *Registry) Convert(source interface{}, targetType reflect.Type) (interface{}, error) {
	// Handle nil values
	if source == nil {
		// Check if target type is nullable
		if isNullableType(targetType) {
			return reflect.Zero(targetType).Interface(), nil
		}
		return nil, fmt.Errorf("cannot convert nil to non-nullable type %v", targetType)
	}

	sourceType := reflect.TypeOf(source)

	// 1. Try exact type match
	if converter, ok := r.converters[TypePair{Source: sourceType, Target: targetType}]; ok {
		return converter(source)
	}

	// 2. Try default converter for target type
	if converter, ok := r.defaults[targetType]; ok {
		return converter(source)
	}

	// 3. Fallback: pass through if types are directly assignable
	if sourceType.AssignableTo(targetType) {
		return source, nil
	}

	// 4. No converter found
	return nil, fmt.Errorf("no converter registered for %v -> %v", sourceType, targetType)
}

// NeedsConversion checks if a conversion is needed for the given type pair
func (r *Registry) NeedsConversion(sourceType, targetType reflect.Type) bool {
	// Types are already compatible
	if sourceType.AssignableTo(targetType) {
		return false
	}

	// Check if we have a converter registered
	_, hasSpecific := r.converters[TypePair{Source: sourceType, Target: targetType}]
	_, hasDefault := r.defaults[targetType]

	return hasSpecific || hasDefault
}

// CreateScanner creates a sql.Scanner that can handle conversion to the target type
func (r *Registry) CreateScanner(targetType reflect.Type) sql.Scanner {
	return &convertingScanner{
		registry:   r,
		targetType: targetType,
	}
}

// convertingScanner implements sql.Scanner and uses the registry for conversion
type convertingScanner struct {
	registry   *Registry
	targetType reflect.Type
	result     interface{}
}

// Scan implements sql.Scanner interface
func (s *convertingScanner) Scan(src interface{}) error {
	result, err := s.registry.Convert(src, s.targetType)
	if err != nil {
		return err
	}
	s.result = result
	return nil
}

// Result returns the converted value
func (s *convertingScanner) Result() interface{} {
	return s.result
}

// Helper functions

// isNullableType checks if a type can hold null values
func isNullableType(t reflect.Type) bool {
	// Check for sql.Null* types
	switch t {
	case reflect.TypeOf(sql.NullBool{}),
		reflect.TypeOf(sql.NullByte{}),
		reflect.TypeOf(sql.NullFloat64{}),
		reflect.TypeOf(sql.NullInt16{}),
		reflect.TypeOf(sql.NullInt32{}),
		reflect.TypeOf(sql.NullInt64{}),
		reflect.TypeOf(sql.NullString{}),
		reflect.TypeOf(sql.NullTime{}):
		return true
	}

	// Check for pointer types
	if t.Kind() == reflect.Ptr {
		return true
	}

	return false
}
