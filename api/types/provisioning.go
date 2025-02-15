/*
Copyright 2020-2022 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/gravitational/trace"
	"golang.org/x/exp/slices"

	"github.com/gravitational/teleport/api/defaults"
	apiutils "github.com/gravitational/teleport/api/utils"
)

// JoinMethod is the method used for new nodes to join the cluster.
type JoinMethod string

const (
	JoinMethodUnspecified JoinMethod = ""
	// JoinMethodToken is the default join method, nodes join the cluster by
	// presenting a secret token.
	JoinMethodToken JoinMethod = "token"
	// JoinMethodEC2 indicates that the node will join with the EC2 join method.
	JoinMethodEC2 JoinMethod = "ec2"
	// JoinMethodIAM indicates that the node will join with the IAM join method.
	JoinMethodIAM JoinMethod = "iam"
	// JoinMethodGitHub indicates that the node will join with the GitHub join
	// method. Documentation regarding the implementation of this can be found
	// in lib/githubactions
	JoinMethodGitHub JoinMethod = "github"
	// JoinMethodCircleCI indicates that the node will join with the CircleCI\
	// join method. Documentation regarding the implementation of this can be
	// found in lib/circleci
	JoinMethodCircleCI JoinMethod = "circleci"
	// JoinMethodKubernetes indicates that the node will join with the
	// Kubernetes join method. Documentation regarding implementation can be
	// found in lib/kubernetestoken
	JoinMethodKubernetes JoinMethod = "kubernetes"
	// JoinMethodAzure indicates that the node will join with the Azure join
	// method.
	JoinMethodAzure JoinMethod = "azure"
	// JoinMethodGitLab indicates that the node will join with the GitLab
	// join method. Documentation regarding implementation of this
	// can be found in lib/gitlab
	JoinMethodGitLab JoinMethod = "gitlab"
	// JoinMethodGCP indicates that the node will join with the GCP join method.
	// Documentation regarding implementation of this can be found in lib/gcp.
	JoinMethodGCP JoinMethod = "gcp"
)

var JoinMethods = []JoinMethod{
	JoinMethodToken,
	JoinMethodEC2,
	JoinMethodIAM,
	JoinMethodGitHub,
	JoinMethodCircleCI,
	JoinMethodKubernetes,
	JoinMethodAzure,
	JoinMethodGitLab,
	JoinMethodGCP,
}

func ValidateJoinMethod(method JoinMethod) error {
	hasJoinMethod := slices.Contains(JoinMethods, method)
	if !hasJoinMethod {
		return trace.BadParameter("join method must be one of %s", apiutils.JoinStrings(JoinMethods, ", "))
	}

	return nil
}

type KubernetesJoinType string

var (
	KubernetesJoinTypeUnspecified KubernetesJoinType = ""
	KubernetesJoinTypeInCluster   KubernetesJoinType = "in_cluster"
	KubernetesJoinTypeStaticJWKS  KubernetesJoinType = "static_jwks"
)

// ProvisionToken is a provisioning token
type ProvisionToken interface {
	ResourceWithOrigin
	// SetMetadata sets resource metatada
	SetMetadata(meta Metadata)
	// GetRoles returns a list of teleport roles
	// that will be granted to the user of the token
	// in the crendentials
	GetRoles() SystemRoles
	// SetRoles sets teleport roles
	SetRoles(SystemRoles)
	// GetAllowRules returns the list of allow rules
	GetAllowRules() []*TokenRule
	// SetAllowRules sets the allow rules
	SetAllowRules([]*TokenRule)
	// GetAWSIIDTTL returns the TTL of EC2 IIDs
	GetAWSIIDTTL() Duration
	// GetJoinMethod returns joining method that must be used with this token.
	GetJoinMethod() JoinMethod
	// GetBotName returns the BotName field which must be set for joining bots.
	GetBotName() string

	// GetSuggestedLabels returns the set of labels that the resource should add when adding itself to the cluster
	GetSuggestedLabels() Labels

	// GetSuggestedAgentMatcherLabels returns the set of labels that should be watched when an agent/service uses this token.
	// An example of this is the Database Agent.
	// When using the install-database.sh script, the script will add those labels as part of the `teleport.yaml` configuration.
	// They are added to `db_service.resources.0.labels`.
	GetSuggestedAgentMatcherLabels() Labels

	// V1 returns V1 version of the resource
	V1() *ProvisionTokenV1
	// String returns user friendly representation of the resource
	String() string

	// GetSafeName returns the name of the token, sanitized appropriately for
	// join methods where the name is secret. This should be used when logging
	// the token name.
	GetSafeName() string
}

// NewProvisionToken returns a new provision token with the given roles.
func NewProvisionToken(token string, roles SystemRoles, expires time.Time) (ProvisionToken, error) {
	return NewProvisionTokenFromSpec(token, expires, ProvisionTokenSpecV2{
		Roles: roles,
	})
}

// NewProvisionTokenFromSpec returns a new provision token with the given spec.
func NewProvisionTokenFromSpec(token string, expires time.Time, spec ProvisionTokenSpecV2) (ProvisionToken, error) {
	t := &ProvisionTokenV2{
		Metadata: Metadata{
			Name:    token,
			Expires: &expires,
		},
		Spec: spec,
	}
	if err := t.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}
	return t, nil
}

// MustCreateProvisionToken returns a new valid provision token
// or panics, used in tests
func MustCreateProvisionToken(token string, roles SystemRoles, expires time.Time) ProvisionToken {
	t, err := NewProvisionToken(token, roles, expires)
	if err != nil {
		panic(err)
	}
	return t
}

// setStaticFields sets static resource header and metadata fields.
func (p *ProvisionTokenV2) setStaticFields() {
	p.Kind = KindToken
	p.Version = V2
}

// CheckAndSetDefaults checks and set default values for any missing fields.
func (p *ProvisionTokenV2) CheckAndSetDefaults() error {
	p.setStaticFields()
	if err := p.Metadata.CheckAndSetDefaults(); err != nil {
		return trace.Wrap(err)
	}

	if len(p.Spec.Roles) == 0 {
		return trace.BadParameter("provisioning token is missing roles")
	}
	roles, err := NewTeleportRoles(SystemRoles(p.Spec.Roles).StringSlice())
	if err != nil {
		return trace.Wrap(err)
	}
	p.Spec.Roles = roles

	if roles.Include(RoleBot) && p.Spec.BotName == "" {
		return trace.BadParameter("token with role %q must set bot_name", RoleBot)
	}

	if p.Spec.BotName != "" && !roles.Include(RoleBot) {
		return trace.BadParameter("can only set bot_name on token with role %q", RoleBot)
	}

	hasAllowRules := len(p.Spec.Allow) > 0
	if p.Spec.JoinMethod == JoinMethodUnspecified {
		// Default to the ec2 join method if any allow rules were specified,
		// else default to the token method. These defaults are necessary for
		// backwards compatibility.
		if hasAllowRules {
			p.Spec.JoinMethod = JoinMethodEC2
		} else {
			p.Spec.JoinMethod = JoinMethodToken
		}
	}
	switch p.Spec.JoinMethod {
	case JoinMethodToken:
		if hasAllowRules {
			return trace.BadParameter("allow rules are not compatible with the %q join method", JoinMethodToken)
		}
	case JoinMethodEC2:
		if !hasAllowRules {
			return trace.BadParameter("the %q join method requires defined token allow rules", JoinMethodEC2)
		}
		for _, allowRule := range p.Spec.Allow {
			if allowRule.AWSARN != "" {
				return trace.BadParameter(`the %q join method does not support the "aws_arn" parameter`, JoinMethodEC2)
			}
			if allowRule.AWSAccount == "" && allowRule.AWSRole == "" {
				return trace.BadParameter(`allow rule for %q join method must set "aws_account" or "aws_role"`, JoinMethodEC2)
			}
		}
		if p.Spec.AWSIIDTTL == 0 {
			// default to 5 minute ttl if unspecified
			p.Spec.AWSIIDTTL = Duration(5 * time.Minute)
		}
	case JoinMethodIAM:
		if !hasAllowRules {
			return trace.BadParameter("the %q join method requires defined token allow rules", JoinMethodIAM)
		}
		for _, allowRule := range p.Spec.Allow {
			if allowRule.AWSRole != "" {
				return trace.BadParameter(`the %q join method does not support the "aws_role" parameter`, JoinMethodIAM)
			}
			if len(allowRule.AWSRegions) != 0 {
				return trace.BadParameter(`the %q join method does not support the "aws_regions" parameter`, JoinMethodIAM)
			}
			if allowRule.AWSAccount == "" && allowRule.AWSARN == "" {
				return trace.BadParameter(`allow rule for %q join method must set "aws_account" or "aws_arn"`, JoinMethodEC2)
			}
		}
	case JoinMethodGitHub:
		providerCfg := p.Spec.GitHub
		if providerCfg == nil {
			return trace.BadParameter(
				`"github" configuration must be provided for join method %q`,
				JoinMethodGitHub,
			)
		}
		if err := providerCfg.checkAndSetDefaults(); err != nil {
			return trace.Wrap(err)
		}
	case JoinMethodCircleCI:
		providerCfg := p.Spec.CircleCI
		if providerCfg == nil {
			return trace.BadParameter(
				`"cirleci" configuration must be provided for join method %q`,
				JoinMethodCircleCI,
			)
		}
		if err := providerCfg.checkAndSetDefaults(); err != nil {
			return trace.Wrap(err)
		}
	case JoinMethodKubernetes:
		providerCfg := p.Spec.Kubernetes
		if providerCfg == nil {
			return trace.BadParameter(
				`"kubernetes" configuration must be provided for the join method %q`,
				JoinMethodKubernetes,
			)
		}
		if err := providerCfg.checkAndSetDefaults(); err != nil {
			return trace.Wrap(err, "spec.kubernetes:")
		}
	case JoinMethodAzure:
		providerCfg := p.Spec.Azure
		if providerCfg == nil {
			return trace.BadParameter(
				`"azure" configuration must be provided for the join method %q`,
				JoinMethodAzure,
			)
		}
		if err := providerCfg.checkAndSetDefaults(); err != nil {
			return trace.Wrap(err)
		}
	case JoinMethodGitLab:
		providerCfg := p.Spec.GitLab
		if providerCfg == nil {
			return trace.BadParameter(
				`"gitlab" configuration must be provided for the join method %q`,
				JoinMethodGitLab,
			)
		}
		if err := providerCfg.checkAndSetDefaults(); err != nil {
			return trace.Wrap(err)
		}
	case JoinMethodGCP:
		providerCfg := p.Spec.GCP
		if providerCfg == nil {
			return trace.BadParameter(
				`"gcp" configuration must be provided for the join method %q`,
				JoinMethodGCP,
			)
		}
		if err := providerCfg.checkAndSetDefaults(); err != nil {
			return trace.Wrap(err)
		}
	default:
		return trace.BadParameter("unknown join method %q", p.Spec.JoinMethod)
	}

	return nil
}

// GetVersion returns resource version
func (p *ProvisionTokenV2) GetVersion() string {
	return p.Version
}

// GetRoles returns a list of teleport roles
// that will be granted to the user of the token
// in the crendentials
func (p *ProvisionTokenV2) GetRoles() SystemRoles {
	// Ensure that roles are case-insensitive.
	return normalizedSystemRoles(SystemRoles(p.Spec.Roles).StringSlice())
}

// SetRoles sets teleport roles
func (p *ProvisionTokenV2) SetRoles(r SystemRoles) {
	p.Spec.Roles = r
}

// GetAllowRules returns the list of allow rules
func (p *ProvisionTokenV2) GetAllowRules() []*TokenRule {
	return p.Spec.Allow
}

// SetAllowRules sets the allow rules.
func (p *ProvisionTokenV2) SetAllowRules(rules []*TokenRule) {
	p.Spec.Allow = rules
}

// GetAWSIIDTTL returns the TTL of EC2 IIDs
func (p *ProvisionTokenV2) GetAWSIIDTTL() Duration {
	return p.Spec.AWSIIDTTL
}

// GetJoinMethod returns joining method that must be used with this token.
func (p *ProvisionTokenV2) GetJoinMethod() JoinMethod {
	return p.Spec.JoinMethod
}

// GetBotName returns the BotName field which must be set for joining bots.
func (p *ProvisionTokenV2) GetBotName() string {
	return p.Spec.BotName
}

// GetKind returns resource kind
func (p *ProvisionTokenV2) GetKind() string {
	return p.Kind
}

// GetSubKind returns resource sub kind
func (p *ProvisionTokenV2) GetSubKind() string {
	return p.SubKind
}

// SetSubKind sets resource subkind
func (p *ProvisionTokenV2) SetSubKind(s string) {
	p.SubKind = s
}

// GetResourceID returns resource ID
func (p *ProvisionTokenV2) GetResourceID() int64 {
	return p.Metadata.ID
}

// SetResourceID sets resource ID
func (p *ProvisionTokenV2) SetResourceID(id int64) {
	p.Metadata.ID = id
}

// GetRevision returns the revision
func (p *ProvisionTokenV2) GetRevision() string {
	return p.Metadata.GetRevision()
}

// SetRevision sets the revision
func (p *ProvisionTokenV2) SetRevision(rev string) {
	p.Metadata.SetRevision(rev)
}

// GetMetadata returns metadata
func (p *ProvisionTokenV2) GetMetadata() Metadata {
	return p.Metadata
}

// SetMetadata sets resource metatada
func (p *ProvisionTokenV2) SetMetadata(meta Metadata) {
	p.Metadata = meta
}

// Origin returns the origin value of the resource.
func (p *ProvisionTokenV2) Origin() string {
	return p.Metadata.Origin()
}

// SetOrigin sets the origin value of the resource.
func (p *ProvisionTokenV2) SetOrigin(origin string) {
	p.Metadata.SetOrigin(origin)
}

// GetSuggestedLabels returns the labels the resource should set when using this token
func (p *ProvisionTokenV2) GetSuggestedLabels() Labels {
	return p.Spec.SuggestedLabels
}

// GetAgentMatcherLabels returns the set of labels that should be watched when an agent/service uses this token.
// An example of this is the Database Agent.
// When using the install-database.sh script, the script will add those labels as part of the `teleport.yaml` configuration.
// They are added to `db_service.resources.0.labels`.
func (p *ProvisionTokenV2) GetSuggestedAgentMatcherLabels() Labels {
	return p.Spec.SuggestedAgentMatcherLabels
}

// V1 returns V1 version of the resource
func (p *ProvisionTokenV2) V1() *ProvisionTokenV1 {
	return &ProvisionTokenV1{
		Roles:   p.Spec.Roles,
		Expires: p.Metadata.Expiry(),
		Token:   p.Metadata.Name,
	}
}

// V2 returns V2 version of the resource
func (p *ProvisionTokenV2) V2() *ProvisionTokenV2 {
	return p
}

// SetExpiry sets expiry time for the object
func (p *ProvisionTokenV2) SetExpiry(expires time.Time) {
	p.Metadata.SetExpiry(expires)
}

// Expiry returns object expiry setting
func (p *ProvisionTokenV2) Expiry() time.Time {
	return p.Metadata.Expiry()
}

// GetName returns the name of the provision token. This value can be secret!
// Use GetSafeName where the name may be logged.
func (p *ProvisionTokenV2) GetName() string {
	return p.Metadata.Name
}

// SetName sets the name of the provision token.
func (p *ProvisionTokenV2) SetName(e string) {
	p.Metadata.Name = e
}

// GetSafeName returns the name of the token, sanitized appropriately for
// join methods where the name is secret. This should be used when logging
// the token name.
func (p *ProvisionTokenV2) GetSafeName() string {
	name := p.GetName()
	if p.GetJoinMethod() != JoinMethodToken {
		return name
	}

	// If the token name is short, we just blank the whole thing.
	if len(name) < 16 {
		return strings.Repeat("*", len(name))
	}

	// If the token name is longer, we can show the last 25% of it to help
	// the operator identify it.
	hiddenBefore := int(0.75 * float64(len(name)))
	name = name[hiddenBefore:]
	name = strings.Repeat("*", hiddenBefore) + name
	return name
}

// String returns the human readable representation of a provisioning token.
func (p ProvisionTokenV2) String() string {
	expires := "never"
	if !p.Expiry().IsZero() {
		expires = p.Expiry().String()
	}
	return fmt.Sprintf("ProvisionToken(Roles=%v, Expires=%v)", p.Spec.Roles, expires)
}

// ProvisionTokensToV1 converts provision tokens to V1 list
func ProvisionTokensToV1(in []ProvisionToken) []ProvisionTokenV1 {
	if in == nil {
		return nil
	}
	out := make([]ProvisionTokenV1, len(in))
	for i := range in {
		out[i] = *in[i].V1()
	}
	return out
}

// ProvisionTokensFromV1 converts V1 provision tokens to resource list
func ProvisionTokensFromV1(in []ProvisionTokenV1) []ProvisionToken {
	if in == nil {
		return nil
	}
	out := make([]ProvisionToken, len(in))
	for i := range in {
		out[i] = in[i].V2()
	}
	return out
}

// V1 returns V1 version of the resource
func (p *ProvisionTokenV1) V1() *ProvisionTokenV1 {
	return p
}

// V2 returns V2 version of the resource
func (p *ProvisionTokenV1) V2() *ProvisionTokenV2 {
	t := &ProvisionTokenV2{
		Kind:    KindToken,
		Version: V2,
		Metadata: Metadata{
			Name:      p.Token,
			Namespace: defaults.Namespace,
		},
		Spec: ProvisionTokenSpecV2{
			Roles: p.Roles,
		},
	}
	if !p.Expires.IsZero() {
		t.SetExpiry(p.Expires)
	}
	t.CheckAndSetDefaults()
	return t
}

// String returns the human readable representation of a provisioning token.
func (p ProvisionTokenV1) String() string {
	expires := "never"
	if p.Expires.Unix() != 0 {
		expires = p.Expires.String()
	}
	return fmt.Sprintf("ProvisionToken(Roles=%v, Expires=%v)",
		p.Roles, expires)
}

func (a *ProvisionTokenSpecV2GitHub) checkAndSetDefaults() error {
	if len(a.Allow) == 0 {
		return trace.BadParameter("the %q join method requires at least one token allow rule", JoinMethodGitHub)
	}
	for _, rule := range a.Allow {
		repoSet := rule.Repository != ""
		ownerSet := rule.RepositoryOwner != ""
		subSet := rule.Sub != ""
		if !(subSet || ownerSet || repoSet) {
			return trace.BadParameter(
				`allow rule for %q must include at least one of "repository", "repository_owner" or "sub"`,
				JoinMethodGitHub,
			)
		}
	}
	if strings.Contains(a.EnterpriseServerHost, "/") {
		return trace.BadParameter("'spec.github.enterprise_server_host' should not contain the scheme or path")
	}
	return nil
}

func (a *ProvisionTokenSpecV2CircleCI) checkAndSetDefaults() error {
	if len(a.Allow) == 0 {
		return trace.BadParameter("the %q join method requires at least one token allow rule", JoinMethodCircleCI)
	}
	if a.OrganizationID == "" {
		return trace.BadParameter("the %q join method requires 'organization_id' to be set", JoinMethodCircleCI)
	}
	for _, rule := range a.Allow {
		projectSet := rule.ProjectID != ""
		contextSet := rule.ContextID != ""
		if !projectSet && !contextSet {
			return trace.BadParameter(
				`allow rule for %q must include at least "project_id" or "context_id"`,
				JoinMethodCircleCI,
			)
		}
	}
	return nil
}

func (a *ProvisionTokenSpecV2Kubernetes) checkAndSetDefaults() error {
	if len(a.Allow) == 0 {
		return trace.BadParameter("allow: at least one rule must be set")
	}
	for i, allowRule := range a.Allow {
		if allowRule.ServiceAccount == "" {
			return trace.BadParameter(
				"allow[%d].service_account: name of service account must be set",
				i,
			)
		}
		if len(strings.Split(allowRule.ServiceAccount, ":")) != 2 {
			return trace.BadParameter(
				`allow[%d].service_account: name of service account should be in format "namespace:service_account", got %q instead`,
				i,
				allowRule.ServiceAccount,
			)
		}
	}

	if a.Type == KubernetesJoinTypeUnspecified {
		// For compatibility with older resources which did not have a Type
		// field we default to "in_cluster".
		a.Type = KubernetesJoinTypeInCluster
	}
	switch a.Type {
	case KubernetesJoinTypeInCluster:
		if a.StaticJWKS != nil {
			return trace.BadParameter("static_jwks: must not be set when type is %q", KubernetesJoinTypeInCluster)
		}
	case KubernetesJoinTypeStaticJWKS:
		if a.StaticJWKS == nil {
			return trace.BadParameter("static_jwks: must be set when type is %q", KubernetesJoinTypeStaticJWKS)
		}
		if a.StaticJWKS.JWKS == "" {
			return trace.BadParameter("static_jwks.jwks: must be set when type is %q", KubernetesJoinTypeStaticJWKS)
		}
	default:
		return trace.BadParameter(
			"type: must be one of (%s), got %q",
			apiutils.JoinStrings(JoinMethods, ", "),
			a.Type,
		)
	}

	return nil
}

func (a *ProvisionTokenSpecV2Azure) checkAndSetDefaults() error {
	if len(a.Allow) == 0 {
		return trace.BadParameter(
			"the %q join method requires defined azure allow rules",
			JoinMethodAzure,
		)
	}
	for _, allowRule := range a.Allow {
		if allowRule.Subscription == "" {
			return trace.BadParameter(
				"the %q join method requires azure allow rules with non-empty subscription",
				JoinMethodAzure,
			)
		}
	}
	return nil
}

const defaultGitLabDomain = "gitlab.com"

func (a *ProvisionTokenSpecV2GitLab) checkAndSetDefaults() error {
	if len(a.Allow) == 0 {
		return trace.BadParameter(
			"the %q join method requires defined gitlab allow rules",
			JoinMethodGitLab,
		)
	}
	for _, allowRule := range a.Allow {
		if allowRule.Sub == "" && allowRule.NamespacePath == "" && allowRule.ProjectPath == "" {
			return trace.BadParameter(
				"the %q join method requires allow rules with at least 'sub', 'project_path' or 'namespace_path' to ensure security.",
				JoinMethodGitLab,
			)
		}
	}

	if a.Domain == "" {
		a.Domain = defaultGitLabDomain
	} else {
		if strings.Contains(a.Domain, "/") {
			return trace.BadParameter(
				"'spec.gitlab.domain' should not contain the scheme or path",
			)
		}
	}
	return nil
}

func (a *ProvisionTokenSpecV2GCP) checkAndSetDefaults() error {
	if len(a.Allow) == 0 {
		return trace.BadParameter("the %q join method requires at least one token allow rule", JoinMethodGCP)
	}
	for _, allowRule := range a.Allow {
		if len(allowRule.ProjectIDs) == 0 {
			return trace.BadParameter(
				"the %q join method requires gcp allow rules with at least one project ID",
				JoinMethodGCP,
			)
		}
	}
	return nil
}
