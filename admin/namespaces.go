// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package admin

import (
	"io"
	"net/http"
	"strings"

	"github.com/maniksurtani/quotaservice/config"
	pb "github.com/maniksurtani/quotaservice/protos/config"
)

type namespacesAPIHandler struct {
	a Administrable
}

func NewNamespacesAPIHandler(admin Administrable) (a *namespacesAPIHandler) {
	return &namespacesAPIHandler{a: admin}
}

func (a *namespacesAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ns := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api"), "/")

	switch r.Method {
	case "GET":
		err := writeNamespace(a, w, ns)

		if err != nil {
			writeJSONError(w, err)
		}
	case "DELETE":
		err := a.a.DeleteNamespace(ns)

		if err != nil {
			writeJSONError(w, &HttpError{err.Error(), http.StatusBadRequest})
		}
	case "PUT":
		changeNamespace(w, r, func(c *pb.NamespaceConfig) error {
			return a.a.UpdateNamespace(c)
		})
	case "POST":
		changeNamespace(w, r, func(c *pb.NamespaceConfig) error {
			return a.a.AddNamespace(c)
		})
	default:
		writeJSONError(w, &HttpError{"Unknown method " + r.Method, http.StatusBadRequest})
	}
}

func writeNamespace(a *namespacesAPIHandler, w http.ResponseWriter, namespace string) *HttpError {
	var object interface{}
	cfgs := a.a.Configs()

	if namespace == "" || namespace == config.GlobalNamespace {
		object = cfgs
	} else {
		if cfgs.Namespaces[namespace] == nil {
			return &HttpError{"Unable to locate namespace " + namespace, http.StatusNotFound}
		}

		object = cfgs.Namespaces[namespace]
	}

	writeJSON(w, object)
	return nil
}

func changeNamespace(w http.ResponseWriter, r *http.Request, updater func(*pb.NamespaceConfig) error) {
	c, e := getNamespaceConfig(r.Body)

	if e != nil {
		writeJSONError(w, &HttpError{e.Error(), http.StatusInternalServerError})
		return
	}

	e = updater(c)

	if e != nil {
		writeJSONError(w, &HttpError{e.Error(), http.StatusInternalServerError})
	}
}

func getNamespaceConfig(r io.Reader) (*pb.NamespaceConfig, error) {
	c := &pb.NamespaceConfig{}
	err := unmarshalJSON(r, c)
	return c, err
}