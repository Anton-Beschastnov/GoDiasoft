package hw09structvalidator

import (
	"errors"
	"testing"
)

type UserRole string

type (
	Meta struct {
		SystemID string `validate:"len:5"`
	}

	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int      `validate:"min:18|max:50"`
		Email  string   `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole `validate:"in:admin,stuff"`
		Phones []string `validate:"len:11"`
		Meta   Meta     `validate:"nested"`
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		in          interface{}
		expectedErr error
	}{
		{
			name: "Valid user structure",
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Age:    25,
				Email:  "test@example.com",
				Role:   "admin",
				Phones: []string{"79991112233", "79991112244"},
				Meta:   Meta{SystemID: "abcde"},
			},
			expectedErr: nil,
		},
		{
			name:        "Not a struct error",
			in:          "just a string",
			expectedErr: ErrNotAStruct,
		},
		{
			name: "Multiple business validation errors",
			in: User{
				ID:    "short-id",
				Age:   15,
				Email: "bad-email",
				Role:  "guest",
				Meta:  Meta{SystemID: "abcde"},
			},
			expectedErr: ErrLengthMismatch,
		},
		{
			name: "Slice elements validation error",
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Age:    30,
				Email:  "test@example.com",
				Role:   "stuff",
				Phones: []string{"79991112233", "123"},
				Meta:   Meta{SystemID: "abcde"},
			},
			expectedErr: ErrLengthMismatch,
		},
		{
			name: "Nested structure error",
			in: User{
				ID:    "12345678-1234-1234-1234-123456789012",
				Age:   30,
				Email: "test@example.com",
				Role:  "stuff",
				Meta:  Meta{SystemID: "too-long"},
			},
			expectedErr: ErrLengthMismatch,
		},
		{
			name: "Invalid tag syntax program error",
			in: struct {
				BadField int `validate:"min"`
			}{BadField: 10},
			expectedErr: ErrInvalidTagFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.in)
			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error %v, got nil", tt.expectedErr)
			}

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error to be or contain %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestAccumulatedErrorsCount(t *testing.T) {
	u := User{
		ID:    "short",
		Age:   99,
		Email: "invalid",
		Role:  "admin",
		Meta:  Meta{SystemID: "abcde"},
	}

	err := Validate(u)
	var valErrs ValidationErrors
	if !errors.As(err, &valErrs) {
		t.Fatalf("expected ValidationErrors type, got %v", err)
	}

	if len(valErrs) != 3 {
		t.Errorf("expected exactly 3 errors, got %d", len(valErrs))
	}
}
