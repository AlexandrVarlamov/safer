package validation

import (
	"strings"
	"unicode/utf8"

	"github.com/go-playground/validator"
)

// New returns a validator with the custom tags used by the API
// ("bucket" and "key") registered on top of the built-in ones.
func New() *validator.Validate {
	v := validator.New()
	_ = v.RegisterValidation("bucket", validBucket)
	_ = v.RegisterValidation("key", validObjectKey)
	return v
}

// validBucket enforces the R2/S3 bucket naming rules: 3–63 characters of
// lowercase letters, digits and hyphens, starting and ending with a letter
// or digit.
func validBucket(fl validator.FieldLevel) bool {
	name := fl.Field().String()

	if l := len(name); l < 3 || l > 63 {
		return false
	}

	for i := 0; i < len(name); i++ {
		c := name[i]
		if !isLowerAlnum(c) && c != '-' {
			return false
		}
	}

	return isLowerAlnum(name[0]) && isLowerAlnum(name[len(name)-1])
}

func isLowerAlnum(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= '0' && c <= '9'
}

// validObjectKey enforces the object key rules we care about: 1–1024 bytes
// of valid UTF-8, no leading slash, no NUL bytes and no empty, "." or ".."
// path segments.
func validObjectKey(fl validator.FieldLevel) bool {
	key := fl.Field().String()

	if !utf8.ValidString(key) {
		return false
	}
	if n := len(key); n == 0 || n > 1024 {
		return false
	}
	if strings.HasPrefix(key, "/") || strings.ContainsRune(key, 0) {
		return false
	}

	for _, seg := range strings.Split(key, "/") {
		if seg == "" || seg == "." || seg == ".." {
			return false
		}
	}

	return true
}
