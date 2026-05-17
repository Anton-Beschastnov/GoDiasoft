package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrNotAStruct           = errors.New("provided value is not a struct")
	ErrInvalidTagFormat     = errors.New("invalid tag format")
	ErrUnsupportedFieldType = errors.New("unsupported field type for this validator")
)

var (
	ErrLengthMismatch = errors.New("length condition not met")
	ErrRegexpMismatch = errors.New("string does not match regular expression")
	ErrNotInSet       = errors.New("value is not in the allowed set")
	ErrValueTooSmall  = errors.New("value is less than minimum")
	ErrValueTooBig    = errors.New("value exceeds maximum")
)

type ValidationError struct {
	Field string
	Err   error
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("field %s: %v", e.Field, e.Err)
}

func (e ValidationError) Is(target error) bool {
	return errors.Is(e.Err, target)
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var sb strings.Builder
	for i, err := range v {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}

func (v ValidationErrors) Is(target error) bool {
	for _, e := range v {
		if errors.Is(e.Err, target) {
			return true
		}
	}
	return false
}

func Validate(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return ErrNotAStruct
	}

	var validationErrs ValidationErrors
	for i := 0; i < val.NumField(); i++ {
		fieldOpts := val.Type().Field(i)
		if !fieldOpts.IsExported() {
			continue
		}
		tag := fieldOpts.Tag.Get("validate")
		if tag == "" {
			continue
		}

		errs, err := validateStructField(val.Field(i), fieldOpts.Name, tag)
		if err != nil {
			return err
		}
		validationErrs = append(validationErrs, errs...)
	}

	if len(validationErrs) > 0 {
		return validationErrs
	}
	return nil
}

func validateStructField(fieldVal reflect.Value, fieldName, tag string) (ValidationErrors, error) {
	if tag == "nested" {
		return validateNested(fieldVal, fieldName)
	}
	return validateRules(fieldVal, fieldName, tag)
}

func validateNested(fieldVal reflect.Value, fieldName string) (ValidationErrors, error) {
	err := Validate(fieldVal.Interface())
	if err == nil {
		return nil, nil
	}
	var nestedErrs ValidationErrors
	if errors.As(err, &nestedErrs) {
		var result ValidationErrors
		for _, nErr := range nestedErrs {
			result = append(result, ValidationError{
				Field: fieldName + "." + nErr.Field,
				Err:   nErr.Err,
			})
		}
		return result, nil
	}
	return nil, err
}

func validateRules(fieldVal reflect.Value, fieldName, tag string) (ValidationErrors, error) {
	var validationErrs ValidationErrors
	rules := strings.Split(tag, "|")
	for _, rule := range rules {
		if err := validateField(fieldVal, rule); err != nil {
			if isProgramError(err) {
				return nil, fmt.Errorf("field %s: %w", fieldName, err)
			}
			validationErrs = append(validationErrs, ValidationError{
				Field: fieldName,
				Err:   err,
			})
		}
	}
	return validationErrs, nil
}

func isProgramError(err error) bool {
	return errors.Is(err, ErrInvalidTagFormat) || errors.Is(err, ErrUnsupportedFieldType)
}

func validateField(val reflect.Value, rule string) error {
	parts := strings.SplitN(rule, ":", 2)
	if len(parts) != 2 {
		return ErrInvalidTagFormat
	}
	command, args := parts[0], parts[1]

	if val.Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			if err := validateValue(val.Index(i), command, args); err != nil {
				return err
			}
		}
		return nil
	}

	return validateValue(val, command, args)
}

func validateValue(val reflect.Value, command, args string) error {
	switch command {
	case "len":
		return validateLen(val, args)
	case "regexp":
		return validateRegexp(val, args)
	case "min":
		return validateMin(val, args)
	case "max":
		return validateMax(val, args)
	case "in":
		return validateIn(val, args)
	default:
		return ErrInvalidTagFormat
	}
}

func validateLen(val reflect.Value, args string) error {
	expectedLen, err := strconv.Atoi(args)
	if err != nil {
		return ErrInvalidTagFormat
	}
	if val.Kind() != reflect.String {
		return ErrUnsupportedFieldType
	}
	if len(val.String()) != expectedLen {
		return ErrLengthMismatch
	}
	return nil
}

func validateRegexp(val reflect.Value, args string) error {
	if val.Kind() != reflect.String {
		return ErrUnsupportedFieldType
	}
	re, err := regexp.Compile(args)
	if err != nil {
		return ErrInvalidTagFormat
	}
	if !re.MatchString(val.String()) {
		return ErrRegexpMismatch
	}
	return nil
}

func validateMin(val reflect.Value, args string) error {
	minVal, err := strconv.Atoi(args)
	if err != nil {
		return ErrInvalidTagFormat
	}
	if val.Kind() != reflect.Int {
		return ErrUnsupportedFieldType
	}
	if int(val.Int()) < minVal {
		return ErrValueTooSmall
	}
	return nil
}

func validateMax(val reflect.Value, args string) error {
	maxVal, err := strconv.Atoi(args)
	if err != nil {
		return ErrInvalidTagFormat
	}
	if val.Kind() != reflect.Int {
		return ErrUnsupportedFieldType
	}
	if int(val.Int()) > maxVal {
		return ErrValueTooBig
	}
	return nil
}

func validateIn(val reflect.Value, args string) error {
	allowed := strings.Split(args, ",")
	switch val.Kind() {
	case reflect.String:
		return validateInString(val.String(), allowed)
	case reflect.Int:
		return validateInInt(int(val.Int()), allowed)
	default:
		return ErrUnsupportedFieldType
	}
}

func validateInString(strVal string, allowed []string) error {
	for _, item := range allowed {
		if strVal == item {
			return nil
		}
	}
	return ErrNotInSet
}

func validateInInt(intVal int, allowed []string) error {
	for _, item := range allowed {
		parsedInt, err := strconv.Atoi(item)
		if err != nil {
			return ErrInvalidTagFormat
		}
		if intVal == parsedInt {
			return nil
		}
	}
	return ErrNotInSet
}
