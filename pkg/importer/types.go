package importer

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/anz-bank/sysl/pkg/syslutil"
	"github.com/sirupsen/logrus"
)

type Func func(args OutputData, text string, logger *logrus.Logger) (out string, err error)

type Type interface {
	Name() string
	Attributes() []string
	AddAttributes([]string) []string
}

type baseType struct {
	name  string
	attrs []string
}

func (s *baseType) Name() string         { return s.name }
func (s *baseType) Attributes() []string { return s.attrs }
func (s *baseType) AddAttributes(attrs []string) []string {
	s.attrs = append(s.attrs, attrs...)
	return s.Attributes()
}

type StandardType struct {
	baseType
	Properties FieldList
}

func (s *StandardType) SortProperties() error {
	// this is done so that sorted properties can be assigned back.
	props, err := s.Properties.SortWithoutDupl()
	if err != nil {
		return err
	}
	s.Properties = props
	return nil
}

// Union represents a union type https://sysl.io/docs/lang-spec#union
type Union struct {
	baseType
	Options FieldList
}

type SyslBuiltIn struct {
	baseType
	name syslutil.BuiltInType
}

func (s *SyslBuiltIn) Name() string { return s.name }

var StringAlias = &SyslBuiltIn{name: syslutil.Type_STRING}

// !alias type without the EXTERNAL_ prefix
type Alias struct {
	baseType
	Target Type
}

type ExternalAlias struct {
	baseType
	Target Type
}

func NewStringAlias(name string, attrs ...string) Type {
	return &ExternalAlias{
		baseType: baseType{
			name:  name,
			attrs: attrs,
		},
		Target: StringAlias,
	}
}

func (s *ExternalAlias) Name() string { return s.name }

type ImportedBuiltInAlias struct {
	baseType // name is the input language type name
	Target   Type
}

type Array struct {
	baseType
	Items Type
}

type Enum struct {
	baseType
}

type maxType int

const (
	MinOnly maxType = iota
	MaxSpecified
	OpenEnded
)

type sizeSpec struct {
	Min     int
	Max     int
	MaxType maxType
}

type Field struct {
	Name     string
	Type     Type
	Optional bool
	Attrs    []string
	SizeSpec *sizeSpec
}

type TypeList struct {
	types []Type
}

func (t TypeList) Items() []Type {
	return t.types
}

func (t TypeList) Sort() {
	sort.SliceStable(t.types, func(i, j int) bool {
		a := t.types[i].Name()
		b := t.types[j].Name()
		return strings.Compare(a, b) < 0
	})
}

type FieldList []Field

func (props FieldList) SortWithoutDupl() (FieldList, error) {
	m := make(map[string]Field)
	for _, p := range props {
		if prop, exist := m[p.Name]; exist && !reflect.DeepEqual(p, prop) {
			return nil, fmt.Errorf("duplicate fields exist: %q", prop.Name)
		}
		m[p.Name] = p
	}
	props = make(FieldList, 0, len(m))
	for _, prop := range m {
		props = append(props, prop)
	}
	sort.SliceStable(props, func(i, j int) bool {
		return strings.Compare(props[i].Name, props[j].Name) < 0
	})
	return props, nil
}

func (t TypeList) Find(name string) (Type, bool) {
	if builtin, ok := checkBuiltInTypes(name); ok {
		return builtin, ok
	}

	for _, n := range t.Items() {
		if n.Name() == name {
			if importAlias, ok := n.(*ImportedBuiltInAlias); ok {
				return importAlias.Target, true
			}
			return n, true
		}
	}
	return &StandardType{}, false
}

func (t *TypeList) Add(item ...Type) {
	for _, i := range item {
		if i.Name() != "" {
			t.types = append(t.types, i)
		}
	}
}

func (t *TypeList) AddAndRet(item Type) Type {
	if item.Name() != "" {
		t.types = append(t.types, item)
	}
	return item
}

func checkBuiltInTypes(name string) (Type, bool) {
	if syslutil.IsBuiltIn(name) {
		return &SyslBuiltIn{name: name}, true
	}
	return &StandardType{}, false
}
