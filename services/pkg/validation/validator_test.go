package validation

import (
	"testing"
)

type TestStruct struct {
	Name     string `validate:"required"`
	Email    string `validate:"required,email"`
	Age      int    `validate:"min=0,max=150"`
	Website  string `validate:"url"`
	Category string `validate:"oneof=gold silver bronze"`
}

func TestValidate_Success(t *testing.T) {
	valid := TestStruct{
		Name:     "John Doe",
		Email:    "john@example.com",
		Age:      30,
		Website:  "https://example.com",
		Category: "gold",
	}

	err := Validate(valid)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidate_RequiredField(t *testing.T) {
	invalid := TestStruct{
		Email:    "john@example.com",
		Age:      30,
		Website:  "https://example.com",
		Category: "gold",
	}

	err := Validate(invalid)
	if err == nil {
		t.Error("expected validation error for missing Name field")
	}
}

func TestValidate_InvalidEmail(t *testing.T) {
	invalid := TestStruct{
		Name:     "John Doe",
		Email:    "not-an-email",
		Age:      30,
		Website:  "https://example.com",
		Category: "gold",
	}

	err := Validate(invalid)
	if err == nil {
		t.Error("expected validation error for invalid email")
	}
}

func TestValidate_MinMax(t *testing.T) {
	invalid := TestStruct{
		Name:     "John Doe",
		Email:    "john@example.com",
		Age:      200,
		Website:  "https://example.com",
		Category: "gold",
	}

	err := Validate(invalid)
	if err == nil {
		t.Error("expected validation error for age > 150")
	}
}

func TestValidate_InvalidURL(t *testing.T) {
	invalid := TestStruct{
		Name:     "John Doe",
		Email:    "john@example.com",
		Age:      30,
		Website:  "not-a-url",
		Category: "gold",
	}

	err := Validate(invalid)
	if err == nil {
		t.Error("expected validation error for invalid URL")
	}
}

func TestValidate_OneOf(t *testing.T) {
	invalid := TestStruct{
		Name:     "John Doe",
		Email:    "john@example.com",
		Age:      30,
		Website:  "https://example.com",
		Category: "platinum",
	}

	err := Validate(invalid)
	if err == nil {
		t.Error("expected validation error for invalid category")
	}
}

func TestMustValidate_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected MustValidate to panic on invalid struct")
		}
	}()

	invalid := TestStruct{
		Email: "not-an-email",
	}

	MustValidate(invalid)
}
