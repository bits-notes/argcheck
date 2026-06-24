package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	maxJSONSchemaDepth      = 8
	maxJSONSchemaProperties = 200
	maxJSONSchemaRequired   = 200
	maxJSONSchemaEnumValues = 200
	maxJSONSchemaPatternLen = 512
)

var supportedJSONSchemaKeys = map[string]bool{
	"type":                 true,
	"required":             true,
	"properties":           true,
	"additionalProperties": true,
	"items":                true,
	"minItems":             true,
	"maxItems":             true,
	"minLength":            true,
	"maxLength":            true,
	"enum":                 true,
	"const":                true,
	"minimum":              true,
	"maximum":              true,
	"exclusiveMinimum":     true,
	"exclusiveMaximum":     true,
	"pattern":              true,
	"title":                true,
	"description":          true,
	"default":              true,
	"examples":             true,
}

type jsonSchemaOptions struct {
	AllowSingleFencedBlock bool `json:"allow_single_fenced_block"`
}

type jsonSchemaValidator struct {
	root    *jsonSchemaNode
	options jsonSchemaOptions
}

type jsonSchemaNode struct {
	types                []string
	required             []string
	properties           map[string]*jsonSchemaNode
	additionalProperties *bool
	items                *jsonSchemaNode
	minItems             *int
	maxItems             *int
	minLength            *int
	maxLength            *int
	enumValues           []any
	constValue           any
	hasConst             bool
	minimum              *float64
	maximum              *float64
	exclusiveMinimum     *float64
	exclusiveMaximum     *float64
	pattern              *regexp.Regexp
}

func compileJSONSchema(spec Spec) (OutputValidator, error) {
	if len(spec.Name) > 0 {
		return nil, fmt.Errorf("json_schema validator does not support name")
	}
	options, err := parseJSONSchemaOptions(spec.Options)
	if err != nil {
		return nil, err
	}
	root, err := parseJSONSchemaNode(spec.Spec, 0)
	if err != nil {
		return nil, err
	}
	return jsonSchemaValidator{root: root, options: options}, nil
}

func parseJSONSchemaOptions(raw json.RawMessage) (jsonSchemaOptions, error) {
	if len(raw) == 0 {
		return jsonSchemaOptions{}, nil
	}
	var options jsonSchemaOptions
	if err := json.Unmarshal(raw, &options); err != nil {
		return jsonSchemaOptions{}, fmt.Errorf("invalid json_schema options: %w", err)
	}
	return options, nil
}

func (v jsonSchemaValidator) Validate(ctx context.Context, output string) Result {
	ctx, cancel := withValidationContext(ctx)
	defer cancel()
	if limit := checkOutputLimit(output); !limit.Passed {
		return limit
	}
	select {
	case <-ctx.Done():
		return Result{Passed: false, Message: "validation timed out"}
	default:
	}

	normalized := strings.ReplaceAll(output, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		return Result{Passed: false, Message: "output is empty"}
	}
	unwrapped, unwrapResult := unwrapJSONOutput(normalized, v.options)
	if !unwrapResult.Passed {
		return unwrapResult
	}

	decoder := json.NewDecoder(strings.NewReader(unwrapped))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return Result{Passed: false, Message: fmt.Sprintf("output is not valid JSON: %v", err)}
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return Result{Passed: false, Message: "output must contain exactly one JSON value"}
	}
	if err := v.root.validate("$", value); err != nil {
		return Result{Passed: false, Message: err.Error()}
	}

	raw, err := canonicalJSON(value)
	if err != nil {
		raw = json.RawMessage(unwrapped)
	}
	return Result{
		Passed:       true,
		Message:      "output matches json_schema",
		ParsedOutput: raw,
	}
}

func unwrapJSONOutput(output string, options jsonSchemaOptions) (string, Result) {
	if !strings.HasPrefix(output, "```") {
		return output, Result{Passed: true}
	}
	if !options.AllowSingleFencedBlock {
		return "", Result{Passed: false, Message: "fenced code blocks are not allowed for this validator"}
	}
	lines := strings.Split(output, "\n")
	if len(lines) < 3 || !fenceLinePattern.MatchString(lines[0]) || strings.TrimSpace(lines[len(lines)-1]) != "```" {
		return "", Result{Passed: false, Message: "output must contain exactly one fenced JSON block"}
	}
	for _, line := range lines[1 : len(lines)-1] {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			return "", Result{Passed: false, Message: "output contains an extra code fence inside the JSON block"}
		}
	}
	content := strings.TrimSpace(strings.Join(lines[1:len(lines)-1], "\n"))
	if content == "" {
		return "", Result{Passed: false, Message: "fenced JSON block is empty"}
	}
	return content, Result{Passed: true}
}

func parseJSONSchemaNode(raw json.RawMessage, depth int) (*jsonSchemaNode, error) {
	if depth > maxJSONSchemaDepth {
		return nil, fmt.Errorf("json_schema exceeds maximum depth %d", maxJSONSchemaDepth)
	}
	if len(raw) == 0 {
		raw = json.RawMessage(`{}`)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var obj map[string]json.RawMessage
	if err := decoder.Decode(&obj); err != nil {
		return nil, fmt.Errorf("invalid json_schema spec: %w", err)
	}
	if obj == nil {
		return nil, fmt.Errorf("json_schema spec must be an object")
	}
	for key := range obj {
		if !supportedJSONSchemaKeys[key] {
			return nil, fmt.Errorf("json_schema keyword %q is not supported", key)
		}
	}

	node := &jsonSchemaNode{}
	var err error
	if rawType, ok := obj["type"]; ok {
		node.types, err = parseJSONSchemaTypes(rawType)
		if err != nil {
			return nil, err
		}
	}
	if rawRequired, ok := obj["required"]; ok {
		node.required, err = parseStringArray(rawRequired, "required")
		if err != nil {
			return nil, err
		}
		if len(node.required) > maxJSONSchemaRequired {
			return nil, fmt.Errorf("json_schema required exceeds %d entries", maxJSONSchemaRequired)
		}
	}
	if rawProperties, ok := obj["properties"]; ok {
		var props map[string]json.RawMessage
		if err := json.Unmarshal(rawProperties, &props); err != nil {
			return nil, fmt.Errorf("json_schema properties must be an object: %w", err)
		}
		if len(props) > maxJSONSchemaProperties {
			return nil, fmt.Errorf("json_schema properties exceeds %d entries", maxJSONSchemaProperties)
		}
		node.properties = make(map[string]*jsonSchemaNode, len(props))
		for name, rawProp := range props {
			child, err := parseJSONSchemaNode(rawProp, depth+1)
			if err != nil {
				return nil, fmt.Errorf("property %q: %w", name, err)
			}
			node.properties[name] = child
		}
	}
	if rawAdditional, ok := obj["additionalProperties"]; ok {
		var value bool
		if err := json.Unmarshal(rawAdditional, &value); err != nil {
			return nil, fmt.Errorf("json_schema additionalProperties must be a boolean")
		}
		node.additionalProperties = &value
	}
	if rawItems, ok := obj["items"]; ok {
		child, err := parseJSONSchemaNode(rawItems, depth+1)
		if err != nil {
			return nil, fmt.Errorf("items: %w", err)
		}
		node.items = child
	}
	if node.minItems, err = parseOptionalInt(obj, "minItems"); err != nil {
		return nil, err
	}
	if node.maxItems, err = parseOptionalInt(obj, "maxItems"); err != nil {
		return nil, err
	}
	if node.minLength, err = parseOptionalInt(obj, "minLength"); err != nil {
		return nil, err
	}
	if node.maxLength, err = parseOptionalInt(obj, "maxLength"); err != nil {
		return nil, err
	}
	if node.minimum, err = parseOptionalNumber(obj, "minimum"); err != nil {
		return nil, err
	}
	if node.maximum, err = parseOptionalNumber(obj, "maximum"); err != nil {
		return nil, err
	}
	if node.exclusiveMinimum, err = parseOptionalNumber(obj, "exclusiveMinimum"); err != nil {
		return nil, err
	}
	if node.exclusiveMaximum, err = parseOptionalNumber(obj, "exclusiveMaximum"); err != nil {
		return nil, err
	}
	if rawPattern, ok := obj["pattern"]; ok {
		var pattern string
		if err := json.Unmarshal(rawPattern, &pattern); err != nil {
			return nil, fmt.Errorf("json_schema pattern must be a string")
		}
		if len(pattern) > maxJSONSchemaPatternLen {
			return nil, fmt.Errorf("json_schema pattern exceeds %d bytes", maxJSONSchemaPatternLen)
		}
		node.pattern, err = regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("json_schema pattern is invalid: %w", err)
		}
	}
	if rawEnum, ok := obj["enum"]; ok {
		var values []json.RawMessage
		if err := json.Unmarshal(rawEnum, &values); err != nil {
			return nil, fmt.Errorf("json_schema enum must be an array")
		}
		if len(values) > maxJSONSchemaEnumValues {
			return nil, fmt.Errorf("json_schema enum exceeds %d values", maxJSONSchemaEnumValues)
		}
		node.enumValues = make([]any, 0, len(values))
		for _, rawValue := range values {
			value, err := decodeJSONValue(rawValue)
			if err != nil {
				return nil, fmt.Errorf("json_schema enum value is invalid: %w", err)
			}
			node.enumValues = append(node.enumValues, value)
		}
	}
	if rawConst, ok := obj["const"]; ok {
		value, err := decodeJSONValue(rawConst)
		if err != nil {
			return nil, fmt.Errorf("json_schema const value is invalid: %w", err)
		}
		node.constValue = value
		node.hasConst = true
	}
	if node.minItems != nil && node.maxItems != nil && *node.minItems > *node.maxItems {
		return nil, fmt.Errorf("json_schema minItems cannot exceed maxItems")
	}
	if node.minLength != nil && node.maxLength != nil && *node.minLength > *node.maxLength {
		return nil, fmt.Errorf("json_schema minLength cannot exceed maxLength")
	}
	return node, nil
}

func parseJSONSchemaTypes(raw json.RawMessage) ([]string, error) {
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		if !validJSONSchemaType(single) {
			return nil, fmt.Errorf("json_schema type %q is not supported", single)
		}
		return []string{single}, nil
	}
	values, err := parseStringArray(raw, "type")
	if err != nil {
		return nil, fmt.Errorf("json_schema type must be a string or array of strings")
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("json_schema type array cannot be empty")
	}
	seen := map[string]bool{}
	for _, value := range values {
		if !validJSONSchemaType(value) {
			return nil, fmt.Errorf("json_schema type %q is not supported", value)
		}
		if seen[value] {
			return nil, fmt.Errorf("json_schema type %q is duplicated", value)
		}
		seen[value] = true
	}
	return values, nil
}

func validJSONSchemaType(value string) bool {
	switch value {
	case "object", "array", "string", "number", "integer", "boolean", "null":
		return true
	default:
		return false
	}
}

func parseStringArray(raw json.RawMessage, field string) ([]string, error) {
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, fmt.Errorf("json_schema %s must be an array of strings", field)
	}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("json_schema %s cannot contain empty strings", field)
		}
	}
	return values, nil
}

func parseOptionalInt(obj map[string]json.RawMessage, key string) (*int, error) {
	raw, ok := obj[key]
	if !ok {
		return nil, nil
	}
	var value int
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, fmt.Errorf("json_schema %s must be an integer", key)
	}
	if value < 0 {
		return nil, fmt.Errorf("json_schema %s cannot be negative", key)
	}
	return &value, nil
}

func parseOptionalNumber(obj map[string]json.RawMessage, key string) (*float64, error) {
	raw, ok := obj[key]
	if !ok {
		return nil, nil
	}
	value, err := decodeJSONNumber(raw)
	if err != nil {
		return nil, fmt.Errorf("json_schema %s must be a number", key)
	}
	return &value, nil
}

func decodeJSONValue(raw json.RawMessage) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func decodeJSONNumber(raw json.RawMessage) (float64, error) {
	value, err := decodeJSONValue(raw)
	if err != nil {
		return 0, err
	}
	number, ok := value.(json.Number)
	if !ok {
		return 0, fmt.Errorf("not a number")
	}
	return jsonNumberFloat(number)
}

func (s *jsonSchemaNode) validate(path string, value any) error {
	if len(s.types) > 0 {
		matches := false
		for _, typ := range s.types {
			if jsonValueMatchesType(value, typ) {
				matches = true
				break
			}
		}
		if !matches {
			return fmt.Errorf("%s has type %s, expected %s", path, jsonTypeName(value), strings.Join(s.types, " or "))
		}
	}
	if s.hasConst && !jsonValuesEqual(value, s.constValue) {
		return fmt.Errorf("%s does not match const", path)
	}
	if len(s.enumValues) > 0 {
		matches := false
		for _, enumValue := range s.enumValues {
			if jsonValuesEqual(value, enumValue) {
				matches = true
				break
			}
		}
		if !matches {
			return fmt.Errorf("%s is not one of the allowed enum values", path)
		}
	}

	switch typed := value.(type) {
	case map[string]any:
		if err := s.validateObject(path, typed); err != nil {
			return err
		}
	case []any:
		if err := s.validateArray(path, typed); err != nil {
			return err
		}
	case string:
		if err := s.validateString(path, typed); err != nil {
			return err
		}
	case json.Number:
		if err := s.validateNumber(path, typed); err != nil {
			return err
		}
	}
	return nil
}

func (s *jsonSchemaNode) validateObject(path string, value map[string]any) error {
	for _, name := range s.required {
		if _, ok := value[name]; !ok {
			return fmt.Errorf("%s is missing required property %q", path, name)
		}
	}
	for name, child := range s.properties {
		if childValue, ok := value[name]; ok {
			if err := child.validate(path+"."+name, childValue); err != nil {
				return err
			}
		}
	}
	if s.additionalProperties != nil && !*s.additionalProperties {
		known := make(map[string]bool, len(s.properties))
		for name := range s.properties {
			known[name] = true
		}
		extras := make([]string, 0)
		for name := range value {
			if !known[name] {
				extras = append(extras, name)
			}
		}
		if len(extras) > 0 {
			sort.Strings(extras)
			return fmt.Errorf("%s contains unsupported property %q", path, extras[0])
		}
	}
	return nil
}

func (s *jsonSchemaNode) validateArray(path string, value []any) error {
	if s.minItems != nil && len(value) < *s.minItems {
		return fmt.Errorf("%s has %d items, minimum is %d", path, len(value), *s.minItems)
	}
	if s.maxItems != nil && len(value) > *s.maxItems {
		return fmt.Errorf("%s has %d items, maximum is %d", path, len(value), *s.maxItems)
	}
	if s.items != nil {
		for i, item := range value {
			if err := s.items.validate(fmt.Sprintf("%s[%d]", path, i), item); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *jsonSchemaNode) validateString(path string, value string) error {
	length := len([]rune(value))
	if s.minLength != nil && length < *s.minLength {
		return fmt.Errorf("%s has length %d, minimum is %d", path, length, *s.minLength)
	}
	if s.maxLength != nil && length > *s.maxLength {
		return fmt.Errorf("%s has length %d, maximum is %d", path, length, *s.maxLength)
	}
	if s.pattern != nil && !s.pattern.MatchString(value) {
		return fmt.Errorf("%s does not match required pattern", path)
	}
	return nil
}

func (s *jsonSchemaNode) validateNumber(path string, value json.Number) error {
	number, err := jsonNumberFloat(value)
	if err != nil {
		return fmt.Errorf("%s is not a finite number", path)
	}
	if s.minimum != nil && number < *s.minimum {
		return fmt.Errorf("%s is %s, minimum is %s", path, value.String(), formatFloat(*s.minimum))
	}
	if s.maximum != nil && number > *s.maximum {
		return fmt.Errorf("%s is %s, maximum is %s", path, value.String(), formatFloat(*s.maximum))
	}
	if s.exclusiveMinimum != nil && number <= *s.exclusiveMinimum {
		return fmt.Errorf("%s is %s, exclusive minimum is %s", path, value.String(), formatFloat(*s.exclusiveMinimum))
	}
	if s.exclusiveMaximum != nil && number >= *s.exclusiveMaximum {
		return fmt.Errorf("%s is %s, exclusive maximum is %s", path, value.String(), formatFloat(*s.exclusiveMaximum))
	}
	return nil
}

func jsonValueMatchesType(value any, typ string) bool {
	switch typ {
	case "object":
		_, ok := value.(map[string]any)
		return ok
	case "array":
		_, ok := value.([]any)
		return ok
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		_, ok := value.(json.Number)
		return ok
	case "integer":
		number, ok := value.(json.Number)
		return ok && jsonNumberIsInteger(number)
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "null":
		return value == nil
	default:
		return false
	}
}

func jsonTypeName(value any) string {
	switch value.(type) {
	case map[string]any:
		return "object"
	case []any:
		return "array"
	case string:
		return "string"
	case json.Number:
		if jsonNumberIsInteger(value.(json.Number)) {
			return "integer"
		}
		return "number"
	case bool:
		return "boolean"
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%T", value)
	}
}

func jsonNumberFloat(value json.Number) (float64, error) {
	number, err := strconv.ParseFloat(value.String(), 64)
	if err != nil || math.IsInf(number, 0) || math.IsNaN(number) {
		return 0, fmt.Errorf("invalid number")
	}
	return number, nil
}

func jsonNumberIsInteger(value json.Number) bool {
	rat, ok := new(big.Rat).SetString(value.String())
	return ok && rat.IsInt()
}

func jsonNumbersEqual(left json.Number, right json.Number) bool {
	leftRat, leftOK := new(big.Rat).SetString(left.String())
	rightRat, rightOK := new(big.Rat).SetString(right.String())
	return leftOK && rightOK && leftRat.Cmp(rightRat) == 0
}

func jsonValuesEqual(left any, right any) bool {
	switch leftValue := left.(type) {
	case json.Number:
		rightValue, ok := right.(json.Number)
		return ok && jsonNumbersEqual(leftValue, rightValue)
	case map[string]any:
		rightValue, ok := right.(map[string]any)
		if !ok || len(leftValue) != len(rightValue) {
			return false
		}
		for key, value := range leftValue {
			rightItem, ok := rightValue[key]
			if !ok || !jsonValuesEqual(value, rightItem) {
				return false
			}
		}
		return true
	case []any:
		rightValue, ok := right.([]any)
		if !ok || len(leftValue) != len(rightValue) {
			return false
		}
		for i := range leftValue {
			if !jsonValuesEqual(leftValue[i], rightValue[i]) {
				return false
			}
		}
		return true
	default:
		return fmt.Sprintf("%#v", left) == fmt.Sprintf("%#v", right)
	}
}

func canonicalJSON(value any) (json.RawMessage, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
