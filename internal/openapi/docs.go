package openapi

import (
	"sync"
)

// doc.go
type Doc struct {
	Summary     string
	Description string
	Request     any
	Response    any
	Status      int
	Headers     []HeaderParam
}

type HeaderParam struct {
	Name        string
	Description string
	Required    bool
}

// doc.go
var (
	handlerDocs = map[string]*Doc{}
	registryMu  sync.RWMutex
)

func register(doc *Doc, method, path string) {
	key := method + " " + path
	registryMu.Lock()
	handlerDocs[key] = doc
	registryMu.Unlock()
}

func getDoc(method, path string) *Doc {
	key := method + " " + path
	registryMu.RLock()
	doc := handlerDocs[key]
	registryMu.RUnlock()
	return doc
}

// package openapi

// import (
// 	"fmt"
// 	"net/http"
// 	"reflect"

// 	"github.com/EfosaE/credora-backend/internal/utils"
// )

// var handlerDocs = map[uintptr]Doc{}

// func WithDocs(doc Doc, h http.HandlerFunc) http.HandlerFunc {
// 	fmt.Println("From doc.go")
// 	utils.PrintJSON(doc)
// 	ptr := reflect.ValueOf(h).Pointer()
// 	handlerDocs[ptr] = doc
// 	return h
// }
