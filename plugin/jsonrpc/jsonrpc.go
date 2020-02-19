// Package jsonrpc provides implementation of a protoc-gen plugin
package jsonrpc

import (
	"fmt"
	"strings"

	"gitlab.forceup.in/dev-proto/common"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
)

func init() {
	generator.RegisterPlugin(New())
}

const (
	httpPkgPath = "net/http"

	jsonrpcPkgPath     = "gitlab.forceup.in/dev-go/gosf/jsonrpc"
	accessPkgPath      = "gitlab.forceup.in/dev-go/gosf/jsonrpc/access"
	protoCommonPkgPath = "gitlab.forceup.in/dev-proto/common"

	commonEmptyType = ".common.Empty"
)

var _ generator.Plugin = (*Plugin)(nil)

// New return a new Plugin instance
func New() *Plugin {
	return &Plugin{}
}

// Plugin for generating webapi codes
type Plugin struct {
	*generator.Generator

	httpPkg        string
	jsonrpcPkg     string
	accessPkg      string
	protoCommonPkg string
}

// Name returns plugin name
func (p *Plugin) Name() string {
	return "jsonrpc"
}

// Init initiates with given *Generator
func (p *Plugin) Init(g *generator.Generator) {
	p.Generator = g
}

// Generate generates output
func (p *Plugin) Generate(fd *generator.FileDescriptor) {
	if len(fd.FileDescriptorProto.Service) == 0 {
		return
	}

	p.httpPkg = string(p.AddImport(httpPkgPath))
	p.jsonrpcPkg = string(p.AddImport(jsonrpcPkgPath))
	p.accessPkg = string(p.AddImport(accessPkgPath))
	p.protoCommonPkg = string(p.AddImport(protoCommonPkgPath))

	p.P("// Reference imports for jsonrpc")
	p.P("var _ ", p.httpPkg, ".ResponseWriter")
	p.P("var _ ", p.jsonrpcPkg, ".Logger")
	p.P("var _ ", p.accessPkg, ".Fetcher")
	p.P("var _ ", p.protoCommonPkg, ".Empty")
	p.P()

	pkgName := fd.GetPackage()

	for _, sd := range fd.FileDescriptorProto.Service {
		p.generateService(pkgName, sd)
	}
}

func (p *Plugin) generateService(pkgName string, sd *descriptor.ServiceDescriptorProto) {
	srvName := generator.CamelCase(sd.GetName())

	var apiVersion string
	var apiPrefix string

	if opts := sd.GetOptions(); opts != nil {
		if ext, _ := proto.GetExtension(opts, common.E_ApiVersion); ext != nil {
			apiVersion = *((ext).(*string))
		}

		if ext, _ := proto.GetExtension(opts, common.E_ApiPrefix); ext != nil {
			apiPrefix = *((ext).(*string))
		}
	}

	if apiVersion != "" && !strings.HasPrefix(apiVersion, "/") {
		apiVersion = "/" + apiVersion
	}

	if apiPrefix == "" {
		apiPrefix = srvName
	}

	if apiPrefix != "" && !strings.HasPrefix(apiPrefix, "/") {
		apiPrefix = "/" + apiPrefix
	}

	prefix := apiVersion + apiPrefix
	p.P("// API prefix for ", srvName, " server")
	p.P(fmt.Sprintf("const JSONRpcAPIPrefixFor%sServer = %q", srvName, prefix))
	p.P()

	p.P("// returns a *jsonrpc.Mux as the api group for ", srvName)
	p.P(fmt.Sprintf("func NewJSONRpcMuxFor%s(logger %s.Logger, srv %sServer) *%s.Mux {", srvName, p.jsonrpcPkg, srvName, p.jsonrpcPkg))
	p.P(fmt.Sprintf("mux := %s.NewMux(JSONRpcAPIPrefixFor%sServer, logger)", p.jsonrpcPkg, srvName))
	p.P()

	for _, md := range sd.GetMethod() {
		methodName := generator.CamelCase(md.GetName())
		p.P(fmt.Sprintf("mux.Handle(\"/%s\", %s(srv))", methodName, jsonrpcMethodHandlerName(srvName, methodName)))
	}

	p.P()
	p.P("return mux")
	p.P("}")
	p.P()

	for _, md := range sd.GetMethod() {
		p.generateServiceMethod(pkgName, srvName, md)
	}
}

func (p *Plugin) generateServiceMethod(pkgName, srvName string, md *descriptor.MethodDescriptorProto) {
	interfaceName := srvName + "Server"
	methodName := generator.CamelCase(md.GetName())

	var grantScope string
	var grantPerm = common.Perm_READ

	if opts := md.GetOptions(); opts != nil {
		if ext, _ := proto.GetExtension(opts, common.E_GrantScope); ext != nil {
			grantScope = *((ext).(*string))
		}

		if ext, _ := proto.GetExtension(opts, common.E_GrantPerm); ext != nil {
			grantPerm = *((ext).(*common.Perm))
		}
	}

	p.P(fmt.Sprintf("func %s(srv %s) %s.HandlerFunc {", jsonrpcMethodHandlerName(srvName, methodName), interfaceName, p.jsonrpcPkg))
	p.P()
	p.P(fmt.Sprintf("return func(rw %s.ResponseWriter, req *%s.Request) error {", p.httpPkg, p.httpPkg))

	if grantScope != "" {
		p.P(fmt.Sprintf("req, ok := %s.CheckAndInjectAccessPerms(req, %q, %d)", p.accessPkg, grantScope, grantPerm))
		p.P(fmt.Sprintf("if !ok { return jsonrpc.NewRPCErrorWithCode(http.StatusUnauthorized, \"unauthorized request, scope=%s, required=%s\") }", grantScope, common.Perm_name[int32(grantPerm)]))
		p.P()
	}

	inputType := generator.CamelCase(md.GetInputType())
	if inputType == commonEmptyType {
		p.P("input := common.EMPTY")
	} else {
		actualType, err := actualInputTypeString(pkgName, inputType)
		if err != nil {
			p.Error(err, "err captured during generating handler for ", pkgName+".", srvName+".", methodName)
		}

		p.P("input := &", actualType, "{}")
		p.P(fmt.Sprintf("if err := %s.DecodeRequest(req, input); err != nil { return err }", p.jsonrpcPkg))
	}
	p.P()

	p.P(fmt.Sprintf("req = %s.InjectHTTPRequest(req)", p.jsonrpcPkg))
	p.P(fmt.Sprintf("out, err := srv.%s(req.Context(), input)", methodName))
	p.P("if err != nil { return err }")
	p.P()

	p.P(fmt.Sprintf("return %s.EncodeResponse(rw, out)", p.jsonrpcPkg))
	p.P("}")
	p.P("}")
	p.P()
}

// GenerateImports generates import statements
func (p *Plugin) GenerateImports(fd *generator.FileDescriptor) {

}

func jsonrpcMethodHandlerName(srvName, methodName string) string {
	return fmt.Sprintf("_jsonrpc_%s_%s_Handler", srvName, methodName)
}

func trimLeftDots(s string) string {
	return strings.TrimLeft(s, ".")
}

func actualInputTypeString(pkgName, raw string) (string, error) {
	pieces := strings.Split(raw, ".")
	if len(pieces) == 0 || len(pieces) != 3 {
		return "", fmt.Errorf("unexpected input type format, raw=%s", raw)
	}

	if pieces[0] != "" {
		return "", fmt.Errorf("unexpected input type format, raw=%s", raw)
	}

	if pieces[1] == pkgName {
		return pieces[2], nil
	}

	return strings.Join(pieces[1:], "."), nil
}
