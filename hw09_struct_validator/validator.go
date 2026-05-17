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
		fieldVal := val.Field(i)

		if !fieldOpts.IsExported() {
			continue
		}

		tag := fieldOpts.Tag.Get("validate")
		if tag == "" {
			continue
		}

		if tag == "nested" {
			if err := Validate(fieldVal.Interface()); err != nil {
				var nestedErrs ValidationErrors
				if errors.As(err, &nestedErrs) {
					for _, nErr := range nestedErrs {
						validationErrs = append(validationErrs, ValidationError{
							Field: fieldOpts.Name + "." + nErr.Field,
							Err:   nErr.Err,
						})
					}
				} else {
					return err
				}
			}
			continue
		}

		rules := strings.Split(tag, "|")
		for _, rule := range rules {
			if err := validateField(fieldVal, rule); err != nil {
				if isProgramError(err) {
					return fmt.Errorf("field %s: %w", fieldOpts.Name, err)
				}
				validationErrs = append(validationErrs, ValidationError{
					Field: fieldOpts.Name,
					Err:   err,
				})
			}
		}
	}

	if len(validationErrs) > 0 {
		return validationErrs
	}

	return nil
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
	kind := val.Kind()

	switch command {
	case "len":
		expectedLen, err := strconv.Atoi(args)
		if err != nil {
			return ErrInvalidTagFormat
		}
		if kind != reflect.String {
			return ErrUnsupportedFieldType
		}
		if len(val.String()) != expectedLen {
			return ErrLengthMismatch
		}

	case "regexp":
		if kind != reflect.String {
			return ErrUnsupportedFieldType
		}
		re, err := regexp.Compile(args)
		if err != nil {
			return ErrInvalidTagFormat
		}
		if !re.MatchString(val.String()) {
			return ErrRegexpMismatch
		}

	case "min":
		minVal, err := strconv.Atoi(args)
		if err != nil {
			return ErrInvalidTagFormat
		}
		if kind != reflect.Int {
			return ErrUnsupportedFieldType
		}
		if int(val.Int()) < minVal {
			return ErrValueTooSmall
		}

	case "max":
		maxVal, err := strconv.Atoi(args)
		if err != nil {
			return ErrInvalidTagFormat
		}
		if kind != reflect.Int {
			return ErrUnsupportedFieldType
		}
		if int(val.Int()) > maxVal {
			return ErrValueTooBig
		}

	case "in":
		allowed := strings.Split(args, ",")
		if kind == reflect.String {
			strVal := val.String()
			for _, item := range allowed {
				if strVal == item {
					return nil
				}
			}
			return ErrNotInSet
		} else if kind == reflect.Int {
			intVal := int(val.Int())
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
		return ErrUnsupportedFieldType

	default:
		return ErrInvalidTagFormat
	}

	return nil
}
