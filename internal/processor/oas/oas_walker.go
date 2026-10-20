package oas

import (
	"strings"

	"github.com/ohler55/ojg/jp"
	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/jputils"
	"github.com/open-resource-discovery/metadata-compactor-golang/internal/common/utils"
)

type OASWalker struct {
	visitor func(jp.Expr, map[string]any)
}

func Walk(document map[string]any, visitor func(jp.Expr, map[string]any)) {
	(&OASWalker{visitor: visitor}).Walk(document)
}

func (self *OASWalker) Walk(document map[string]any) {
	self.visit(document, jputils.Expr("$"))
	self.info(document, jputils.Expr("$", "info"))
	self.paths(document, jputils.Expr("$", "paths"))
	self.visit(document, jputils.Expr("$", "externalDocs"))
	self.components(document, jputils.Expr("$", "components"))
	self.iterateArray(document, jputils.Expr("$", "tags"), self.tag)
	self.iterateArray(document, jputils.Expr("$", "servers"), self.server)
	self.iterateObject(document, jputils.Expr("$", "webhooks"), self.pathItem)
}

func (self *OASWalker) visit(document map[string]any, path jp.Expr) {
	self.visitor(path, self.objectAt(document, path))
}

func (self *OASWalker) objectAt(document map[string]any, path jp.Expr) map[string]any {
	return utils.SafeCast[map[string]any](path.First(document))
}

func (self *OASWalker) arrayAt(document map[string]any, path jp.Expr) []any {
	return utils.SafeCast[[]any](path.First(document))
}

func (self *OASWalker) iterateArray(document map[string]any, path jp.Expr, walk func(map[string]any, jp.Expr)) {
	for index := range self.arrayAt(document, path) {
		walk(document, path.Nth(index))
	}
}

func (self *OASWalker) iterateObject(document map[string]any, path jp.Expr, walk func(map[string]any, jp.Expr)) {
	for name := range self.objectAt(document, path) {
		walk(document, path.Child(name))
	}
}

func (self *OASWalker) info(document map[string]any, path jp.Expr) {
	self.visit(document, path)
	self.visit(document, path.Child("contact"))
	self.visit(document, path.Child("license"))
}

func (self *OASWalker) server(document map[string]any, path jp.Expr) {
	self.visit(document, path)
	self.iterateObject(document, path.Child("variables"), self.visit)
}

func (self *OASWalker) paths(document map[string]any, path jp.Expr) {
	self.visit(document, path)
	for name := range self.objectAt(document, path) {
		if strings.HasPrefix(name, "/") {
			self.pathItem(document, path.Child(name))
		}
	}
}

func (self *OASWalker) pathItem(document map[string]any, path jp.Expr) {
	self.visit(document, path)
	self.operation(document, path.Child("get"))
	self.operation(document, path.Child("put"))
	self.operation(document, path.Child("post"))
	self.operation(document, path.Child("head"))
	self.operation(document, path.Child("patch"))
	self.operation(document, path.Child("trace"))
	self.operation(document, path.Child("query"))
	self.operation(document, path.Child("delete"))
	self.operation(document, path.Child("options"))
	self.iterateArray(document, path.Child("servers"), self.server)
	self.iterateArray(document, path.Child("parameters"), self.parameter)
	self.iterateObject(document, path.Child("additionalOperations"), self.visit)
}

func (self *OASWalker) operation(document map[string]any, path jp.Expr) {
	self.visit(document, path)
	self.visit(document, path.Child("externalDocs"))
	self.responses(document, path.Child("responses"))
	self.callbackMap(document, path.Child("callbacks"))
	self.requestBody(document, path.Child("requestBody"))
	self.iterateArray(document, path.Child("servers"), self.server)
	self.iterateArray(document, path.Child("parameters"), self.parameter)
}

func (self *OASWalker) parameter(document map[string]any, path jp.Expr) {
	self.visit(document, path)
	self.iterateObject(document, path.Child("examples"), self.visit)
	self.iterateObject(document, path.Child("content"), self.mediaType)
}

func (self *OASWalker) requestBody(document map[string]any, path jp.Expr) {
	self.visit(document, path)
	self.iterateObject(document, path.Child("content"), self.mediaType)
}

func (self *OASWalker) responses(document map[string]any, path jp.Expr) {
	self.visit(document, path)
	for name := range self.objectAt(document, path) {
		if !strings.HasPrefix(name, "x-") {
			self.response(document, path.Child(name))
		}
	}
}

func (self *OASWalker) response(document map[string]any, path jp.Expr) {
	self.visit(document, path)
	self.iterateObject(document, path.Child("links"), self.link)
	self.iterateObject(document, path.Child("headers"), self.header)
	self.iterateObject(document, path.Child("content"), self.mediaType)
}

func (self *OASWalker) header(document map[string]any, path jp.Expr) {
	self.visit(document, path)
	self.iterateObject(document, path.Child("examples"), self.visit)
	self.iterateObject(document, path.Child("content"), self.mediaType)
}

func (self *OASWalker) mediaType(document map[string]any, path jp.Expr) {
	self.visit(document, path)
	self.encoding(document, path.Child("itemEncoding"))
	self.iterateObject(document, path.Child("examples"), self.visit)
	self.iterateObject(document, path.Child("encoding"), self.encoding)
	self.iterateArray(document, path.Child("prefixEncoding"), self.encoding)
}

func (self *OASWalker) encoding(document map[string]any, path jp.Expr) {
	if !path.Has(document) {
		return
	}

	self.visit(document, path)
	self.encoding(document, path.Child("itemEncoding"))
	self.iterateObject(document, path.Child("headers"), self.header)
	self.iterateObject(document, path.Child("encoding"), self.encoding)
	self.iterateArray(document, path.Child("prefixEncoding"), self.encoding)
}

func (self *OASWalker) link(document map[string]any, path jp.Expr) {
	self.visit(document, path)
	self.visit(document, path.Child("server"))
	self.iterateObject(document, path.Child("server").Child("variables"), self.visit)
}

func (self *OASWalker) callbackMap(document map[string]any, path jp.Expr) {
	self.iterateObject(document, path, func(document map[string]any, callback jp.Expr) {
		self.visit(document, callback)
		for expression := range self.objectAt(document, callback) {
			if !strings.HasPrefix(expression, "x-") {
				self.pathItem(document, callback.Child(expression))
			}
		}
	})
}

func (self *OASWalker) tag(document map[string]any, path jp.Expr) {
	self.visit(document, path)
	self.visit(document, path.Child("externalDocs"))
}

func (self *OASWalker) components(document map[string]any, path jp.Expr) {
	self.visit(document, path)
	self.callbackMap(document, path.Child("callbacks"))
	self.iterateObject(document, path.Child("links"), self.link)
	self.iterateObject(document, path.Child("headers"), self.header)
	self.iterateObject(document, path.Child("examples"), self.visit)
	self.iterateObject(document, path.Child("pathItems"), self.pathItem)
	self.iterateObject(document, path.Child("responses"), self.response)
	self.iterateObject(document, path.Child("parameters"), self.parameter)
	self.iterateObject(document, path.Child("mediaTypes"), self.mediaType)
	self.iterateObject(document, path.Child("requestBodies"), self.requestBody)
	self.iterateObject(document, path.Child("securitySchemes"), self.securityScheme)
}

func (self *OASWalker) securityScheme(document map[string]any, path jp.Expr) {
	self.visit(document, path)
	self.visit(document, path.Child("flows"))
	self.visit(document, path.Child("flows").Child("implicit"))
	self.visit(document, path.Child("flows").Child("password"))
	self.visit(document, path.Child("flows").Child("authorizationCode"))
	self.visit(document, path.Child("flows").Child("clientCredentials"))
	self.visit(document, path.Child("flows").Child("deviceAuthorization"))
}
