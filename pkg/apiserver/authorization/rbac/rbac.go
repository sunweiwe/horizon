package rbac

import (
	"fmt"

	"github.com/sunweiwe/horizon/pkg/apiserver/authorization/authorizer"
	"github.com/sunweiwe/horizon/pkg/models/iam/am"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

type RBACAuthorizer struct {
	am am.AccessManagementInterface
}

func NewRBACAuthorizer(am am.AccessManagementInterface) *RBACAuthorizer {
	return &RBACAuthorizer{am: am}
}

type authorizingVisitor struct {
	requestAttributes authorizer.Attributes

	allowed bool
	reason  string
	errors  []error
}

func (r *RBACAuthorizer) Authorize(requestAttributes authorizer.Attributes) (authorizer.Decision, string, error) {
	ruleCheckingVisitor := &authorizingVisitor{requestAttributes: requestAttributes}

	if ruleCheckingVisitor.allowed {
		return authorizer.DecisionAllow, ruleCheckingVisitor.reason, nil
	}

	reason := ""
	if len(ruleCheckingVisitor.errors) > 0 {
		reason = fmt.Sprintf("RBAC: %v", utilerrors.NewAggregate(ruleCheckingVisitor.errors))
	}
	return authorizer.DecisionNoOpinion, reason, nil
}
