package request

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/sunweiwe/horizon/pkg/api"
	"k8s.io/apimachinery/pkg/api/validation/path"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"

	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metainternalversionscheme "k8s.io/apimachinery/pkg/apis/meta/internalversion/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8srequest "k8s.io/apiserver/pkg/endpoints/request"
)

const (
	VerbCreate = "create"
	VerbGet    = "get"
	VerbList   = "list"
	VerbUpdate = "update"
	VerbDelete = "delete"
	VerbWatch  = "watch"
	VerbPatch  = "patch"
)

var kubernetesAPIPrefixes = sets.New("api", "apis")
var specialVerbs = sets.New("proxy", "watch")
var namespaceSubResources = sets.New("status", "finalize")
var specialVerbsNoSubResources = sets.New("proxy")

type RequestInfoResolver interface {
	NewRequestInfo(req *http.Request) (*RequestInfo, error)
}

type RequestInfo struct {
	*k8srequest.RequestInfo

	KubernetesRequest bool

	Workspace string

	Cluster string

	DevOps string

	ResourceScope string

	SourceIP string

	UserAgent string
}

type RequestInfoFactory struct {
	APIPrefixes          sets.Set[string]
	GroupLessAPIPrefixes sets.Set[string]
	GlobalResources      []schema.GroupResource
}

type requestInfoKeyType int

const requestInfoKey requestInfoKeyType = iota

func RequestInfoFrom(ctx context.Context) (*RequestInfo, bool) {
	info, ok := ctx.Value(requestInfoKey).(*RequestInfo)
	return info, ok
}

func WithRequestInfo(parent context.Context, info *RequestInfo) context.Context {
	return k8srequest.WithValue(parent, requestInfoKey, info)
}

func (r *RequestInfoFactory) NewRequestInfo(req *http.Request) (*RequestInfo, error) {
	requestInfo := RequestInfo{
		KubernetesRequest: false,
		RequestInfo: &k8srequest.RequestInfo{
			Path: req.URL.Path,
			Verb: req.Method,
		},
		Workspace: api.ClusterNone,
		Cluster:   api.ClusterNone,
		// SourceIP: ip,
		UserAgent: req.UserAgent(),
	}

	defer func() {
		prefix := requestInfo.APIPrefix
		if prefix == "" {
			currentParts := splitPath(requestInfo.Path)
			if len(currentParts) > 0 && len(currentParts) < 3 {
				prefix = currentParts[0]
			}
		}
		if kubernetesAPIPrefixes.Has(prefix) {
			requestInfo.KubernetesRequest = true
		}
	}()

	currentParts := splitPath(req.URL.Path)
	if len(currentParts) < 3 {
		return &requestInfo, nil
	}

	if !r.APIPrefixes.Has(currentParts[0]) {
		return &requestInfo, nil
	}
	requestInfo.APIPrefix = currentParts[0]
	currentParts = currentParts[1:]

	if currentParts[0] == "clusters" {
		if len(currentParts) > 1 {
			requestInfo.Cluster = currentParts[1]
		}
		if len(currentParts) > 2 {
			currentParts = currentParts[2:]
		}
	}

	if !r.GroupLessAPIPrefixes.Has(requestInfo.APIPrefix) {
		if len(currentParts) < 3 {
			return &requestInfo, nil
		}

		requestInfo.APIGroup = currentParts[0]
		currentParts = currentParts[1:]
	}

	requestInfo.IsResourceRequest = true
	requestInfo.APIVersion = currentParts[0]
	currentParts = currentParts[1:]

	if len(currentParts) > 0 && specialVerbs.Has(currentParts[0]) {
		if len(currentParts) < 2 {
			return &requestInfo, fmt.Errorf("unable to determine kind and namespace from url: %v", req.URL)
		}

		requestInfo.Verb = currentParts[0]
		currentParts = currentParts[1:]
	} else {
		switch req.Method {
		case "POST":
			requestInfo.Verb = VerbCreate
		case "GET", "HEAD":
			requestInfo.Verb = VerbGet
		case "PUT":
			requestInfo.Verb = VerbUpdate
		case "PATCH":
			requestInfo.Verb = VerbPatch
		case "DELETE":
			requestInfo.Verb = VerbDelete
		default:
			requestInfo.Verb = ""
		}
	}

	if currentParts[0] == "workspaces" {
		if len(currentParts) > 1 {
			requestInfo.Workspace = currentParts[1]
		}
		if len(currentParts) > 2 {
			currentParts = currentParts[2:]
		}
	}

	if currentParts[0] == "namespaces" {
		if len(currentParts) > 1 {
			requestInfo.Namespace = currentParts[1]

			if len(currentParts) > 2 && !namespaceSubResources.Has(currentParts[2]) {
				currentParts = currentParts[2:]
			}
		}
	} else if currentParts[0] == "devops" {
		if len(currentParts) > 1 {
			requestInfo.DevOps = currentParts[1]

			if len(currentParts) > 2 {
				currentParts = currentParts[2:]
			}
		}
	} else {
		requestInfo.Namespace = metav1.NamespaceNone
		requestInfo.DevOps = metav1.NamespaceNone
	}

	requestInfo.Parts = currentParts

	switch {
	case len(requestInfo.Parts) >= 3 && !specialVerbsNoSubResources.Has(requestInfo.Verb):
		requestInfo.Subresource = requestInfo.Parts[2]
		fallthrough
	case len(requestInfo.Parts) >= 2:
		requestInfo.Name = requestInfo.Parts[1]
		fallthrough
	case len(requestInfo.Parts) >= 1:
		requestInfo.Resource = requestInfo.Parts[0]
	}

	requestInfo.ResourceScope = r.resolveResourceScope(requestInfo)

	if len(requestInfo.Name) == 0 && requestInfo.Verb == VerbGet {
		opts := metainternalversion.ListOptions{}
		if err := metainternalversionscheme.ParameterCodec.DecodeParameters(req.URL.Query(), metav1.SchemeGroupVersion, &opts); err != nil {
			klog.Errorf("Couldn't parse request %#v: %v", req.URL.Query(), err)

			opts = metainternalversion.ListOptions{}
			if values := req.URL.Query()["watch"]; len(values) > 0 {
				switch strings.ToLower(values[0]) {
				case "false", "0":
				default:
					opts.Watch = true
				}
			}
		}

		if opts.Watch {
			requestInfo.Verb = VerbWatch
		} else {
			requestInfo.Verb = VerbList
		}

		if opts.FieldSelector != nil {
			if name, ok := opts.FieldSelector.RequiresExactMatch("metadata.name"); ok {
				if len(path.IsValidPathSegmentName(name)) == 0 {
					requestInfo.Name = name
				}
			}

		}
	}

	if requestInfo.Verb == VerbWatch {
		selector := req.URL.Query().Get("labelSelector")
		if strings.HasPrefix(selector, workspaceSelectorPrefix) {
			workspace := strings.TrimPrefix(selector, workspaceSelectorPrefix)
			requestInfo.Workspace = workspace
			requestInfo.ResourceScope = WorkspaceScope

		}
	}

	if len(requestInfo.Name) == 0 && requestInfo.Verb == VerbDelete {
		requestInfo.Verb = "deletecollection"
	}

	return &requestInfo, nil
}

func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}

const (
	GlobalScope             = "Global"
	ClusterScope            = "Cluster"
	WorkspaceScope          = "Workspace"
	NamespaceScope          = "Namespace"
	DevOpsScope             = "DevOps"
	workspaceSelectorPrefix = "horizon.io/workspace="
)

func (r *RequestInfoFactory) resolveResourceScope(request RequestInfo) string {
	if r.isGlobalScopeResource(request.APIGroup, request.Resource) {
		return GlobalScope
	}

	if request.Namespace != "" {
		return NamespaceScope
	}

	if request.DevOps != "" {
		return DevOpsScope
	}

	if request.Workspace != "" {
		return WorkspaceScope
	}

	return ClusterScope
}

func (r *RequestInfoFactory) isGlobalScopeResource(apiGroup, resource string) bool {
	for _, groupResource := range r.GlobalResources {
		if groupResource.Group == apiGroup && groupResource.Resource == resource {
			return true
		}
	}
	return false
}
