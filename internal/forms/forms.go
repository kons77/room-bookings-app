package forms

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
)

// Form creates a custom form struct, embeds a url.Values object
type Form struct {
	url.Values
	Errors errors
}

// Valid returns true if there are no errrors, oterwise return false
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// New initializes a form struct
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
		// make(errors), // Fixed: Using make() to initialize the map
	}
}

// Required checks for required fields
func (f *Form) Required(fields ...string) {
	// ... string is a variadic function
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

// Has checks if form field is in post and not empty - for checkbox, etc
func (f *Form) Has(field string) bool {
	x := f.Get(field) // r.Form.Get(field) - a mistake! Have to get field from reciever
	/* if x == "" {
		return false
	}
	return true */
	return x != ""
}

// MinLength checks for string minimum length
func (f *Form) MinLength(field string, length int) bool {
	x := f.Get(field) // r.Form.Get(field) - a mistake! Have to get field from reciever
	if len(x) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long", length))
		return false
	}
	return true
}

// IsEmail checks for valid email address
func (f *Form) IsEmail(field string) {
	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "Invalid email address")
	}
}
