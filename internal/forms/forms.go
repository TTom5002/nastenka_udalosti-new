package forms

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
)

// TODO: Změň komentáře

// Form creates a custom form struct, embeds a url.Values object
type Form struct {
	url.Values
	Errors errors
}

// Valid returns true if there are no errors, otherwise false
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// New initialize a form struct
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

// Required checks for required fields
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "Nutno zadat")
		}
	}
}

// Has checks if form field is in post and not empty
func (f *Form) Has(field string) bool {
	x := f.Get(field)
	if x == "" {
		return false
	}
	return true
}

// MinLength checks for string minimum length
func (f *Form) MinLength(field string, lenght int) bool {
	x := f.Get(field)
	if len(x) < lenght {
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long", lenght))
		return false
	}
	return true
}

// IsEmail checks for valid email address
func (f *Form) IsEmail(field string) {
	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "Neplatná emailová adresa")
	}
}

func (f *Form) SamePassword(password, passwordver string) bool {
	if password != passwordver {
		f.Errors.Add("password", "Hesla se neshodují")
		f.Errors.Add("passwordver", "Hesla se neshodují")
		return false
	}
	return true
}
