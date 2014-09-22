// © 2013 Steve McCoy under the MIT license.

/*
Package validate provides a type for automatically validating the fields of structs.

Any fields tagged with the key "validate" will be validated via a user-defined list of functions.
For example:

	type X struct {
		A string `validate:"long"`
		B string `validate:"short"`
		C string `validate:"long,proper"`
		D string
	}

Multiple validators can be named in the tag by separating their names with commas.
The validators are defined in a map like so:

	vd := make(validate.V)
	vd["long"] = func(i interface{}) error {
		…
	}
	vd["short"] = func(i interface{}) error {
		…
	}
	…

When present in a field's tag, the Validate method passes to these functions the value in the field
and should return an error when the value is deemed invalid.

There is a reserved tag, "struct", which can be used to automatically validate a
struct field, either named or embedded. This may be combined with user-defined validators.

Reflection is used to access the tags and fields, so the usual caveats and limitations apply.
*/
package validate

import (
	"fmt"
	"reflect"
	"strings"
)

// V is a map of tag names to validators.
type V map[string]func(interface{}) error

// Validate accepts a struct (or a pointer) and returns a list of errors for all
// fields that are invalid. If all fields are valid, or s is not a struct type,
// Validate returns nil.
//
// Fields that are not tagged or cannot be interfaced via reflection
// are skipped.
func (v V) Validate(s interface{}) map[string]interface{} {
	errors := make(map[string]interface{})
	v.validate(errors, s)
	if len(errors) > 0 {
		return errors
	}
	return nil
}

func (v V) validate(errs map[string]interface{}, s interface{}) {
	val := reflect.ValueOf(s)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	t := val.Type()
	if t == nil || t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := val.Field(i)
		if !fv.CanInterface() {
			continue
		}
		val := fv.Interface()
		tag := f.Tag.Get("validate")
		if tag == "" {
			continue
		}
		vts := strings.Split(tag, ",")

		for _, vt := range vts {
			if vt == "struct" {
				errs2 := v.Validate(val)
				if errs2 != nil {
					/* A field validation has failed */
					errs[f.Name] = errs2
					break
				}
				continue
			}

			vf := v[vt]
			if vf == nil {
				errs[f.Name] = fmt.Errorf("undefined validator: %q", vt)
				break
			}
			if err := vf(val); err != nil {
				errs[f.Name] = err
				break
			}
		}
	}
}
