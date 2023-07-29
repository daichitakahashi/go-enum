package gen

import (
	"fmt"
	"strings"

	"github.com/IGLOU-EU/go-wildcard"
)

type NamingVisitorParams struct {
	Target     string
	TypeName   string
	MethodName string
}

type NamingAcceptParams struct {
	Target     string
	MethodName string
}

type namingRegistry struct {
	visitors []NamingVisitorParams
	accepts  []NamingAcceptParams

	visitorParamsCache map[string]*NamingVisitorParams
	visitorTypeCache   map[string]string
	visitMethodCache   map[string]string
	acceptMethodCache  map[string]string
}

func newNamingRegistry(visitors []NamingVisitorParams, accepts []NamingAcceptParams) *namingRegistry {
	return &namingRegistry{
		visitors: visitors,
		accepts:  accepts,

		visitorParamsCache: map[string]*NamingVisitorParams{},
		visitorTypeCache:   map[string]string{},
		visitMethodCache:   map[string]string{},
		acceptMethodCache:  map[string]string{}, // identEnum to accept method name
	}
}

func (r *namingRegistry) namingVisitorParams(enumIdent string) (*NamingVisitorParams, bool) {
	if params, ok := r.visitorParamsCache[enumIdent]; ok {
		return params, params != nil
	}

	for _, v := range r.visitors {
		if wildcard.MatchSimple(v.Target, enumIdent) {
			r.visitorParamsCache[enumIdent] = &v
			return &v, true
		}
	}
	r.visitorParamsCache[enumIdent] = nil
	return nil, false
}

func (r *namingRegistry) visitorTypeName(enumIdent string) string {
	if name, ok := r.visitorTypeCache[enumIdent]; ok {
		return name
	}

	pattern := "*Visitor"
	if params, ok := r.namingVisitorParams(enumIdent); ok {
		pattern = params.TypeName
	}
	name := strings.Replace(pattern, "*", enumIdent, 1)
	r.visitorTypeCache[enumIdent] = name
	return name
}

func (r *namingRegistry) visitMethodName(enumIdent, memberName string) string {
	key := fmt.Sprintf("%s:%s", enumIdent, memberName)
	if name, ok := r.visitMethodCache[key]; ok {
		return name
	}

	pattern := "Visit*"
	if params, ok := r.namingVisitorParams(enumIdent); ok {
		pattern = params.MethodName
	}
	name := strings.Replace(pattern, "*", memberName, 1)
	r.visitorTypeCache[key] = name
	return name
}

func (r *namingRegistry) acceptMethodName(enumIdent string) string {
	if name, ok := r.acceptMethodCache[enumIdent]; ok {
		return name
	}

	var namingParams *NamingAcceptParams
	for _, a := range r.accepts {
		if wildcard.MatchSimple(a.Target, enumIdent) {
			namingParams = &a
			break
		}
	}
	if namingParams == nil {
		r.acceptMethodCache[enumIdent] = "Accept"
		return "Accept"
	}
	name := strings.Replace(namingParams.MethodName, "*", enumIdent, 1)
	r.acceptMethodCache[enumIdent] = name
	return name
}
