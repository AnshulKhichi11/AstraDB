package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// Schema for collection
type Schema struct {
	Version    int                    `json:"version"`
	Fields     map[string]FieldSchema `json:"fields"`
	Strict     bool                   `json:"strict"`
	CreatedAt  time.Time              `json:"createdAt"`
	UpdatedAt  time.Time              `json:"updatedAt"`
}

// FieldSchema defines rules for a field
type FieldSchema struct {
	Type      string                 `json:"type"`      // string, number, boolean, object, array
	Required  bool                   `json:"required"`
	MinLength *int                   `json:"minLength,omitempty"`
	MaxLength *int                   `json:"maxLength,omitempty"`
	Min       *float64               `json:"min,omitempty"`
	Max       *float64               `json:"max,omitempty"`
	Pattern   string                 `json:"pattern,omitempty"`
	Format    string                 `json:"format,omitempty"` // email, url
	Enum      []interface{}          `json:"enum,omitempty"`
	Properties map[string]FieldSchema `json:"properties,omitempty"`
	Items     *FieldSchema           `json:"items,omitempty"`
}

// ValidationResult
type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
}

// Validate document
func (s *Schema) Validate(doc map[string]interface{}) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	// Check required fields
	for fieldName, fieldSchema := range s.Fields {
		value, exists := doc[fieldName]

		if fieldSchema.Required && !exists {
			result.Valid = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("Field '%s' is required", fieldName))
			continue
		}

		if !exists {
			continue
		}

		// Validate field
		if err := s.validateField(fieldName, value, fieldSchema); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, err.Error())
		}
	}

	// Strict mode: check extra fields
	if s.Strict {
		for fieldName := range doc {
			if fieldName == "_id" {
				continue
			}
			if _, exists := s.Fields[fieldName]; !exists {
				result.Valid = false
				result.Errors = append(result.Errors,
					fmt.Sprintf("Field '%s' not allowed in strict mode", fieldName))
			}
		}
	}

	return result
}

// Validate individual field
func (s *Schema) validateField(name string, value interface{}, schema FieldSchema) error {
	actualType := getType(value)
	if actualType != schema.Type {
		return fmt.Errorf("Field '%s': expected %s, got %s",
			name, schema.Type, actualType)
	}

	switch schema.Type {
	case "string":
		return s.validateString(name, value.(string), schema)
	case "number":
		// Handle both int and float64
		var numVal float64
		switch v := value.(type) {
		case float64:
			numVal = v
		case int:
			numVal = float64(v)
		case int64:
			numVal = float64(v)
		}
		return s.validateNumber(name, numVal, schema)
	case "object":
		return s.validateObject(name, value.(map[string]interface{}), schema)
	case "array":
		return s.validateArray(name, value.([]interface{}), schema)
	}

	return nil
}

// Validate string
func (s *Schema) validateString(name, value string, schema FieldSchema) error {
	if schema.MinLength != nil && len(value) < *schema.MinLength {
		return fmt.Errorf("Field '%s': minimum length is %d", name, *schema.MinLength)
	}
	if schema.MaxLength != nil && len(value) > *schema.MaxLength {
		return fmt.Errorf("Field '%s': maximum length is %d", name, *schema.MaxLength)
	}

	if schema.Pattern != "" {
		matched, _ := regexp.MatchString(schema.Pattern, value)
		if !matched {
			return fmt.Errorf("Field '%s': does not match pattern", name)
		}
	}

	if schema.Format != "" {
		if err := s.validateFormat(name, value, schema.Format); err != nil {
			return err
		}
	}

	if len(schema.Enum) > 0 {
		found := false
		for _, enumVal := range schema.Enum {
			if value == enumVal {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Field '%s': must be one of %v", name, schema.Enum)
		}
	}

	return nil
}

// Validate number
func (s *Schema) validateNumber(name string, value float64, schema FieldSchema) error {
	if schema.Min != nil && value < *schema.Min {
		return fmt.Errorf("Field '%s': must be >= %f", name, *schema.Min)
	}
	if schema.Max != nil && value > *schema.Max {
		return fmt.Errorf("Field '%s': must be <= %f", name, *schema.Max)
	}
	return nil
}

// Validate object
func (s *Schema) validateObject(name string, value map[string]interface{}, schema FieldSchema) error {
	if schema.Properties == nil {
		return nil
	}

	for propName, propSchema := range schema.Properties {
		propValue, exists := value[propName]
		if propSchema.Required && !exists {
			return fmt.Errorf("Field '%s.%s' is required", name, propName)
		}
		if exists {
			if err := s.validateField(fmt.Sprintf("%s.%s", name, propName), propValue, propSchema); err != nil {
				return err
			}
		}
	}
	return nil
}

// Validate array
func (s *Schema) validateArray(name string, value []interface{}, schema FieldSchema) error {
	if schema.Items == nil {
		return nil
	}

	for i, item := range value {
		if err := s.validateField(fmt.Sprintf("%s[%d]", name, i), item, *schema.Items); err != nil {
			return err
		}
	}
	return nil
}

// Validate format
func (s *Schema) validateFormat(name, value, format string) error {
	switch format {
	case "email":
		pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		matched, _ := regexp.MatchString(pattern, value)
		if !matched {
			return fmt.Errorf("Field '%s': invalid email format", name)
		}
	case "url":
		pattern := `^https?://`
		matched, _ := regexp.MatchString(pattern, value)
		if !matched {
			return fmt.Errorf("Field '%s': invalid URL format", name)
		}
	}
	return nil
}

// Get type of value
func getType(value interface{}) string {
	switch value.(type) {
	case string:
		return "string"
	case float64, int, int64:
		return "number"
	case bool:
		return "boolean"
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}

// Save schema to file
func (e *Engine) SaveSchema(db, collection string, schema *Schema) error {
	collPath := collectionDir(e.cfg, db, collection)
	if err := os.MkdirAll(collPath, 0755); err != nil {
		return err
	}

	schemaPath := filepath.Join(collPath, "schema.json")
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(schemaPath, data, 0644)
}

// Load schema from file
func (e *Engine) LoadSchema(db, collection string) (*Schema, error) {
	schemaPath := filepath.Join(collectionDir(e.cfg, db, collection), "schema.json")
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, err
	}

	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, err
	}

	return &schema, nil
}

// Delete schema
func (e *Engine) DeleteSchema(db, collection string) error {
	schemaPath := filepath.Join(collectionDir(e.cfg, db, collection), "schema.json")
	return os.Remove(schemaPath)
}