package structil

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"text/tabwriter"
	"unsafe"
)

type Getter interface {
	GetRT(name string) reflect.Type
	Has(name string) bool
	Get(name string) interface{}
	GetBytes(name string) []byte
	GetString(name string) string
	GetInt64(name string) int64
	GetUint64(name string) uint64
	GetFloat64(name string) float64
	GetBool(name string) bool
	IsBytes(name string) bool
	IsString(name string) bool
	IsInt64(name string) bool
	IsUint64(name string) bool
	IsFloat64(name string) bool
	IsBool(name string) bool
	IsMap(name string) bool
	IsFunc(name string) bool
	IsChan(name string) bool
	IsStruct(name string) bool
	IsSlice(name string) bool
	MapGet(name string, f func(int, Getter) interface{}) ([]interface{}, error)
	DumpRVs() error
}

type gImpl struct {
	rv        reflect.Value // Value of input interface
	cachedHas map[string]bool
	cachedRT  map[string]reflect.Type  // Type map of struct fields
	cachedRV  map[string]reflect.Value // Value map of indirected struct fields
	cachedI   map[string]interface{}   // interface map of struct fields
}

func NewGetter(i interface{}) (Getter, error) {
	if i == nil {
		return nil, fmt.Errorf("value of passed argument %+v is nil", i)
	}

	rv := reflect.ValueOf(i)
	kind := rv.Kind()

	if kind != reflect.Ptr && kind != reflect.Struct {
		return nil, fmt.Errorf("%v is not supported kind", kind)
	}

	if kind == reflect.Ptr {
		if rv.IsNil() {
			return nil, fmt.Errorf("value of passed argument %+v is nil", rv)
		}

		// indirect is required when kind is Ptr
		rv = reflect.Indirect(rv)
	}

	return &gImpl{
		rv:        rv,
		cachedHas: map[string]bool{},
		cachedRV:  map[string]reflect.Value{},
		cachedRT:  map[string]reflect.Type{},
		cachedI:   map[string]interface{}{},
	}, nil
}

func (g *gImpl) GetRT(name string) reflect.Type {
	_, ok := g.cachedRT[name]
	if !ok {
		g.cache(name)
	}

	return g.cachedRT[name]
}

func (g *gImpl) cache(name string) {
	frv := g.rv.FieldByName(name) // XXX: This code is slow
	if frv.IsValid() {
		g.cachedRT[name] = frv.Type()
		g.cachedHas[name] = true
	} else {
		g.cachedRT[name] = nil
		g.cachedHas[name] = false
	}

	frv = reflect.Indirect(frv)
	g.cachedRV[name] = frv

	g.cachedI[name] = toI(frv)
}

func (g *gImpl) Has(name string) bool {
	_, ok := g.cachedHas[name]
	if !ok {
		g.cache(name)
	}

	return g.cachedHas[name]
}

func (g *gImpl) Get(name string) interface{} {
	_, ok := g.cachedI[name]
	if !ok {
		g.cache(name)
	}

	return g.cachedI[name]
}

func (g *gImpl) getRV(name string) reflect.Value {
	_, ok := g.cachedRV[name]
	if !ok {
		g.cache(name)
	}

	return g.cachedRV[name]
}

func (g *gImpl) GetBytes(name string) []byte {
	return g.getRV(name).Bytes()
}

func (g *gImpl) GetString(name string) string {
	// Note:
	// reflect.Value has String() method because it implements the Stringer interface.
	// So this method does not occur panic.
	return g.getRV(name).String()
}

func (g *gImpl) GetInt64(name string) int64 {
	return g.getRV(name).Int()
}

func (g *gImpl) GetUint64(name string) uint64 {
	return g.getRV(name).Uint()
}

func (g *gImpl) GetFloat64(name string) float64 {
	return g.getRV(name).Float()
}

func (g *gImpl) GetBool(name string) bool {
	return g.getRV(name).Bool()
}

func (g *gImpl) IsBytes(name string) bool {
	return g.IsSlice(name) && g.GetRT(name).Elem().Kind() == reflect.Uint8
}

func (g *gImpl) IsString(name string) bool {
	return g.is(name, reflect.String)
}

func (g *gImpl) IsInt64(name string) bool {
	return g.is(name, reflect.Int64)
}

func (g *gImpl) IsUint64(name string) bool {
	return g.is(name, reflect.Uint64)
}

func (g *gImpl) IsFloat64(name string) bool {
	return g.is(name, reflect.Float64)
}

func (g *gImpl) IsBool(name string) bool {
	return g.is(name, reflect.Bool)
}

func (g *gImpl) IsMap(name string) bool {
	return g.is(name, reflect.Map)
}

func (g *gImpl) IsFunc(name string) bool {
	return g.is(name, reflect.Func)
}

func (g *gImpl) IsChan(name string) bool {
	return g.is(name, reflect.Chan)
}

func (g *gImpl) IsStruct(name string) bool {
	return g.is(name, reflect.Struct)
}

func (g *gImpl) IsSlice(name string) bool {
	return g.is(name, reflect.Slice)
}

func (g *gImpl) is(name string, exp reflect.Kind) bool {
	frv := g.getRV(name)
	return frv.Kind() == exp
}

func (g *gImpl) MapGet(name string, f func(int, Getter) interface{}) ([]interface{}, error) {
	if !g.IsSlice(name) {
		return nil, fmt.Errorf("field %s is not slice", name)
	}

	var vi reflect.Value
	var ac Getter
	var err error
	var res []interface{}
	srv := g.getRV(name)

	for i := 0; i < srv.Len(); i++ {
		vi = srv.Index(i)
		ac, err = NewGetter(toI(vi))
		if err != nil {
			res = append(res, nil)
			continue
		}

		res = append(res, f(i, ac))
	}

	return res, nil
}

func (g *gImpl) DumpRVs() error {
	var rvs []reflect.Value

	for _, rv := range g.cachedRV {
		rvs = append(rvs, rv)
	}

	return dumpValues(rvs)
}

// TODO: candidates of moving to utils
func toI(rv reflect.Value) interface{} {
	if rv.IsValid() && rv.CanInterface() {
		return rv.Interface()
	} else {
		return nil
	}
}

// TODO: candidates of moving to utils
func clone(i interface{}) interface{} {
	return reflect.Indirect(reflect.ValueOf(i)).Interface()
}

// TODO: candidates of moving to utils
func newSettable(typ reflect.Type) reflect.Value {
	return reflect.New(typ).Elem()
}

// TODO: candidates of moving to utils
func settableOf(i interface{}) reflect.Value {
	// i's Kind must be Interface or Ptr(if else, occur panic)
	return reflect.ValueOf(i).Elem()
}

// TODO: candidates of moving to utils
func genericsTypeOf() reflect.Type {
	// generics type is interface pointer
	return reflect.TypeOf((*interface{})(nil)).Elem()
}

// TODO: candidates of moving to utils
func newGenericsSettable() reflect.Value {
	return newSettable(genericsTypeOf())
}

// TODO: candidates of moving to utils
func unexportedField(i interface{}, name string) reflect.Value {
	sv := settableOf(i)
	f := sv.FieldByName(name)
	return reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
}

// TODO: candidates of moving to utils
func compareStructure(i1 interface{}, i2 interface{}) bool {
	// TODO: 2つのstructの構造比較
	return false
}

// TODO: candidates of moving to utils
func dumpValues(rvs []reflect.Value) error {
	var t interface{}
	ds := make([][]interface{}, len(rvs))

	for i, rv := range rvs {
		if rv.IsValid() {
			t = rv.Type()
		} else {
			t = rv.Kind()
		}
		ds[i] = []interface{}{
			t,  // Type
			rv, // Value
		}
	}

	w := getValueWriter(nil)
	w.Write([]byte(fmt.Sprintf("%s\t%s\n", "Type", "Value")))
	w.Write([]byte(fmt.Sprintf("%s\t%s\n", "-----", "-----")))

	for _, d := range ds {
		w.Write([]byte(fmt.Sprintf(
			"%v\t%+v\n", d[0], d[1],
		)))
	}
	err := w.Flush()

	return err
}

// TODO: candidates of moving to utils
func getValueWriter(wrap io.Writer) *tabwriter.Writer {
	if wrap == nil {
		wrap = os.Stdout
	}

	return tabwriter.NewWriter(wrap, 0, 4, 4, ' ', 0)
}