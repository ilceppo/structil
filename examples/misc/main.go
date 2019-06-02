package main

import (
	"io"
	"log"
	"os"
	"reflect"

	"github.com/goldeneggg/structil"
)

type A struct {
	ID       int64
	Name     string
	NamePtr  *string
	IsMan    bool
	FloatVal float64
	AaPtr    *AA
	Nil      *AA
	XArr     []X
	XPtrArr  []*X
	StrArr   []string
}

type AA struct {
	Name   string
	Writer io.Writer
	AaaPtr *AAA
}

type AAA struct {
	Name string
	Val  int
}

type X struct {
	Key   string
	Value string
}

var (
	name = "ほげ　ふがお"

	hoge = &A{
		ID:       1,
		Name:     name,
		NamePtr:  &name,
		IsMan:    true,
		FloatVal: 3.14,
		AaPtr: &AA{
			Name:   "あいう　えおあ",
			Writer: os.Stdout,
			AaaPtr: &AAA{
				Name: "かきく　けこか",
				Val:  8,
			},
		},
		Nil: nil,
		XArr: []X{
			{
				Key:   "key1",
				Value: "value1",
			},
			{
				Key:   "key2",
				Value: "value2",
			},
		},
		XPtrArr: []*X{
			{
				Key:   "key100",
				Value: "value100",
			},
			{
				Key:   "key200",
				Value: "value200",
			},
		},
		StrArr: []string{"key1", "value1", "key2", "value2"},
	}
)

func main() {
	exampleAccessor()
	exampleRetriever()
}

func exampleAccessor() {
	log.Println("---------- exampleAccessor")
	ac, err := structil.NewAccessor(hoge)
	if err != nil {
		log.Printf("!!! ERROR: %v", err)
	}

	name := ac.Get("Name")
	log.Printf("Accessor.Get(Name): %s", name)

	name = ac.GetString("NamePtr")
	log.Printf("Accessor.GetString(NamePtr): %s", name)

	IsMan := ac.GetBool("IsMan")
	log.Printf("Accessor.GetBool(IsMan): %v", IsMan)

	floatVal := ac.GetFloat64("FloatVal")
	log.Printf("Accessor.GetFloat64(FloatVal): %v", floatVal)

	// AaPtr
	aaPtr := ac.Get("AaPtr")
	log.Printf("Accessor.Get(AaPtr): %v", aaPtr)
	log.Printf("Accessor.IsStruct(AaPtr): %v", ac.IsStruct("AaPtr"))
	log.Printf("Accessor.IsInterface(AaPtr): %v", ac.IsInterface("AaPtr"))

	aaAc, err := structil.NewAccessor(aaPtr)
	if err != nil {
		log.Printf("!!! ERROR: %v", err)
	}

	it := aaAc.Get("Writer")
	log.Printf("AaPtr.Get(Writer): %+v", it)
	log.Printf("AaPtr.Get(Writer).ValueOf().Elem(): %+v", reflect.ValueOf(it).Elem())
	log.Printf("AaPtr.IsStruct(Writer): %v", aaAc.IsStruct("Writer"))
	log.Printf("AaPtr.IsInterface(Writer): %v", aaAc.IsInterface("Writer"))

	// Nil
	rvNil := ac.GetRV("Nil")
	log.Printf("Accessor.GetRV(Nil): %v", rvNil)
	aNil := ac.Get("Nil")
	log.Printf("Accessor.Get(Nil): %v", aNil)
	log.Printf("Accessor.IsStruct(Nil): %v", ac.IsStruct("Nil"))
	log.Printf("Accessor.IsInterface(Nil): %v", ac.IsInterface("Nil"))

	aNilAc, err := structil.NewAccessor(aNil)
	if err != nil {
		log.Printf("!!! ERROR: %v", err)
	}
	log.Printf("Accessor.Get(Nil).NewAccessor: %+v", aNilAc)

	// XArr
	xArr := ac.Get("XArr")
	log.Printf("Accessor.Get(XArr): %v", xArr)
	log.Printf("Accessor.IsStruct(XArr): %v", ac.IsStruct("XArr"))
	log.Printf("Accessor.IsSlice(XArr): %v", ac.IsSlice("XArr"))
	log.Printf("Accessor.IsInterface(XArr): %v", ac.IsInterface("XArr"))

	// Map
	fa := func(i int, a structil.Accessor) interface{} {
		s1 := a.GetString("Key")
		s2 := a.GetString("Value")
		return s1 + "=" + s2
	}

	results, err := ac.MapStructs("XArr", fa)
	if err != nil {
		log.Printf("!!! ERROR: %+v", err)
	}
	log.Printf("results XArr: %v, err: %v", results, err)

	results, err = ac.MapStructs("XPtrArr", fa)
	if err != nil {
		log.Printf("!!! ERROR: %+v", err)
	}
	log.Printf("results XPtrArr: %v, err: %v", results, err)
}

func exampleRetriever() {
	log.Println("---------- exampleRetriever")
	ac, err := structil.NewAccessor(hoge)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	swRes, err := structil.NewRetriever().
		Nest("AaPtr").Want("Name").
		Nest("AaaPtr").Want("Name").Want("Val").
		From(hoge)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
	log.Printf("Retriever.From res: %#v", swRes)

	swRes, err = structil.NewRetriever().
		Nest("AaPtr").Want("Name").
		Nest("AaaPtr").Want("Name").Want("Val").
		FromAccessor(ac)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
	log.Printf("Retriever.FromAccessor res: %#v", swRes)
}
