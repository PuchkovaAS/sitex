package validator

import (
	"strings"

	"github.com/gobuffalo/validate"
)

func FormatErrors(errors *validate.Errors) string {
	var res string
	for _, value := range errors.Errors {
		res = res + strings.Join(value, ",") + ", "
	}
	return res
}
