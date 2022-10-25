package cmp

import (
	"reflect"
	"strings"
)

const (
	ROOT_KEY      = "root"
	TAG_OPTION_ID = "id"
)

var fieldNameTags = []string{"json", "yaml"}

type CompareFunc func(*DiffNode, interface{}, reflect.Value, reflect.Value) error

type Comparator struct {
	TagName           string
	RespectSliceOrder bool
	SkipPrivateFields bool
}

func NewComparator() *Comparator {
	return &Comparator{
		TagName: "cmp",
	}
}

// Compare initializes default comparator, compares given values
// and returns a comparison tree.
func Compare(a, b interface{}) (*DiffNode, error) {
	c := NewComparator()
	return c.Compare(a, b)
}

// Compare compares the given elements of the same type and returns
// a comparison tree.
func (c *Comparator) Compare(a, b interface{}) (*DiffNode, error) {
	diff := NewNode()

	if a == nil && b == nil {
		return diff, nil
	}

	err := c.compare(diff, ROOT_KEY, reflect.ValueOf(a), reflect.ValueOf(b))
	return diff.getChild(ROOT_KEY), err
}

// compare recursively compares given values.
func (c *Comparator) compare(parent *DiffNode, key interface{}, a, b reflect.Value) error {
	cmpFunc := c.getCompareFunc(a, b)

	if cmpFunc == nil {
		return NewTypeMismatchError(a.Kind(), b.Kind())
	}

	return cmpFunc(parent, key, a, b)
}

// getCompareFunc returns a compare function based on the type of a
// comparative values.
func (c *Comparator) getCompareFunc(a, b reflect.Value) CompareFunc {
	switch {
	case areOfKind(a, b, reflect.Invalid, reflect.Bool):
		return c.cmpBool
	case areOfKind(a, b, reflect.Invalid, reflect.Int):
		return c.cmpInt
	case areOfKind(a, b, reflect.Invalid, reflect.String):
		return c.cmpString
	case areOfKind(a, b, reflect.Invalid, reflect.Struct):
		return c.cmpStruct
	case areOfKind(a, b, reflect.Invalid, reflect.Slice):
		return c.cmpSlice
	case areOfKind(a, b, reflect.Invalid, reflect.Map):
		return c.cmpMap
	case areOfKind(a, b, reflect.Invalid, reflect.Pointer):
		return c.cmpPointer
	case areOfKind(a, b, reflect.Invalid, reflect.Interface):
		return c.cmpInterface
	default:
		return nil
	}
}

// areComparativeById returns true if one of the values contains a
// tag option representing an ID element.
func (c *Comparator) areComparativeById(a, b reflect.Value) bool {
	if a.Len() > 0 {
		ai := a.Index(0)
		av := getDeepValue(ai)

		if av.Kind() == reflect.Struct {
			if tagOptionId(c.TagName, av) != nil {
				return true
			}
		}
	}

	if b.Len() > 0 {
		bi := b.Index(0)
		bv := getDeepValue(bi)

		if bv.Kind() == reflect.Struct {
			if tagOptionId(c.TagName, bv) != nil {
				return true
			}
		}
	}

	return false
}

func areOfKind(a, b reflect.Value, kinds ...reflect.Kind) bool {
	var isA, isB bool

	for _, k := range kinds {
		if a.Kind() == k {
			isA = true
		}

		if b.Kind() == k {
			isB = true
		}
	}

	return isA && isB
}

func tagName(tagName string, field reflect.StructField) string {
	tags := append([]string{tagName}, fieldNameTags...)

	for _, tag := range tags {
		tName := strings.SplitN(field.Tag.Get(tag), ",", 2)[0]

		if len(tName) > 0 {
			return tName
		}
	}

	return ""
}

func tagOptionId(tagName string, v reflect.Value) interface{} {
	if v.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < v.NumField(); i++ {
		if hasTagOption(tagName, v.Type().Field(i), TAG_OPTION_ID) {
			return exportInterface(v.Field(i))
		}
	}

	return nil
}

func hasTagOption(tagName string, field reflect.StructField, option string) bool {
	tag := field.Tag.Get(tagName)
	options := strings.Split(tag, ",")

	if len(options) < 2 {
		return false
	}

	for _, o := range options[1:] {
		o = strings.TrimSpace(o)
		o = strings.ToLower(o)

		if o == option {
			return true
		}
	}

	return false
}
