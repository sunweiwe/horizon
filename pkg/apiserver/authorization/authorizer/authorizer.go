package authorizer

import "k8s.io/apiserver/pkg/authentication/user"

type Decision int

const (
	DecisionDeny Decision = iota

	DecisionAllow

	DecisionNoOpinion
)

type AtrributesRecord struct {
	User            user.Info
	Verb            string
	Cluster         string
	Workspace       string
	Namespace       string
	DevOps          string
	APIGroup        string
	APIVersion      string
	Resource        string
	Subresource     string
	Name            string
	Path            string
	ResourceScope   string
	ResourceRequest bool
}

type Attributes interface {
}

type Authorizer interface {
	Authorize(attr Attributes) (authorized Decision, reason string, err error)
}

type AuthorizerFunc func(attr Attributes) (Decision, string, error)

func (f AuthorizerFunc) Authorize(attr Attributes) (Decision, string, error) {
	return f(attr)
}
