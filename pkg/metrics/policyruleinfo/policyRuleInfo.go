package policyruleinfo

import (
	"fmt"

	kyverno "github.com/kyverno/kyverno/api/kyverno/v1"
	"github.com/kyverno/kyverno/pkg/autogen"
	"github.com/kyverno/kyverno/pkg/metrics"
	"github.com/kyverno/kyverno/pkg/utils"
	prom "github.com/prometheus/client_golang/prometheus"
)

func (pc PromConfig) registerPolicyRuleInfoMetric(
	policyValidationMode metrics.PolicyValidationMode,
	policyType metrics.PolicyType,
	policyBackgroundMode metrics.PolicyBackgroundMode,
	policyNamespace, policyName, ruleName string,
	ruleType metrics.RuleType,
	metricChangeType PolicyRuleInfoMetricChangeType,
	ready bool,
) error {
	var metricValue float64
	switch metricChangeType {
	case PolicyRuleCreated:
		metricValue = float64(1)
	case PolicyRuleDeleted:
		metricValue = float64(0)
	default:
		return fmt.Errorf("unknown metric change type found:  %s", metricChangeType)
	}

	includeNamespaces, excludeNamespaces := pc.Config.GetIncludeNamespaces(), pc.Config.GetExcludeNamespaces()
	if (policyNamespace != "" && policyNamespace != "-") && utils.ContainsString(excludeNamespaces, policyNamespace) {
		pc.Log.Info(fmt.Sprintf("Skipping the registration of kyverno_policy_rule_info_total metric as the operation belongs to the namespace '%s' which is one of 'namespaces.exclude' %+v in values.yaml", policyNamespace, excludeNamespaces))
		return nil
	}
	if (policyNamespace != "" && policyNamespace != "-") && len(includeNamespaces) > 0 && !utils.ContainsString(includeNamespaces, policyNamespace) {
		pc.Log.Info(fmt.Sprintf("Skipping the registration of kyverno_policy_rule_info_total metric as the operation belongs to the namespace '%s' which is not one of 'namespaces.include' %+v in values.yaml", policyNamespace, includeNamespaces))
		return nil
	}

	if policyType == metrics.Cluster {
		policyNamespace = "-"
	}

	status := "false"
	if ready {
		status = "true"
	}

	pc.Metrics.PolicyRuleInfo.With(prom.Labels{
		"policy_validation_mode": string(policyValidationMode),
		"policy_type":            string(policyType),
		"policy_background_mode": string(policyBackgroundMode),
		"policy_namespace":       policyNamespace,
		"policy_name":            policyName,
		"rule_name":              ruleName,
		"rule_type":              string(ruleType),
		"status_ready":           status,
	}).Set(metricValue)

	return nil
}

func (pc PromConfig) AddPolicy(policy interface{}) error {
	switch inputPolicy := policy.(type) {
	case *kyverno.ClusterPolicy:
		policyValidationMode, err := metrics.ParsePolicyValidationMode(inputPolicy.Spec.GetValidationFailureAction())
		if err != nil {
			return err
		}
		policyBackgroundMode := metrics.ParsePolicyBackgroundMode(inputPolicy)
		policyType := metrics.Cluster
		policyNamespace := "" // doesn't matter for cluster policy
		policyName := inputPolicy.GetName()
		ready := inputPolicy.IsReady()
		// registering the metrics on a per-rule basis
		for _, rule := range autogen.ComputeRules(inputPolicy) {
			ruleName := rule.Name
			ruleType := metrics.ParseRuleType(rule)

			if err = pc.registerPolicyRuleInfoMetric(policyValidationMode, policyType, policyBackgroundMode, policyNamespace, policyName, ruleName, ruleType, PolicyRuleCreated, ready); err != nil {
				return err
			}
		}
		return nil
	case *kyverno.Policy:
		policyValidationMode, err := metrics.ParsePolicyValidationMode(inputPolicy.Spec.GetValidationFailureAction())
		if err != nil {
			return err
		}
		policyBackgroundMode := metrics.ParsePolicyBackgroundMode(inputPolicy)
		policyType := metrics.Namespaced
		policyNamespace := inputPolicy.GetNamespace()
		policyName := inputPolicy.GetName()
		ready := inputPolicy.IsReady()
		// registering the metrics on a per-rule basis
		for _, rule := range autogen.ComputeRules(inputPolicy) {
			ruleName := rule.Name
			ruleType := metrics.ParseRuleType(rule)

			if err = pc.registerPolicyRuleInfoMetric(policyValidationMode, policyType, policyBackgroundMode, policyNamespace, policyName, ruleName, ruleType, PolicyRuleCreated, ready); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("wrong input type provided %T. Only kyverno.Policy and kyverno.ClusterPolicy allowed", inputPolicy)
	}
}

func (pc PromConfig) RemovePolicy(policy interface{}) error {
	switch inputPolicy := policy.(type) {
	case *kyverno.ClusterPolicy:
		for _, rule := range autogen.ComputeRules(inputPolicy) {
			policyValidationMode, err := metrics.ParsePolicyValidationMode(inputPolicy.Spec.GetValidationFailureAction())
			if err != nil {
				return err
			}
			policyBackgroundMode := metrics.ParsePolicyBackgroundMode(inputPolicy)
			policyType := metrics.Cluster
			policyNamespace := "" // doesn't matter for cluster policy
			policyName := inputPolicy.GetName()
			ruleName := rule.Name
			ruleType := metrics.ParseRuleType(rule)
			ready := inputPolicy.IsReady()

			if err = pc.registerPolicyRuleInfoMetric(policyValidationMode, policyType, policyBackgroundMode, policyNamespace, policyName, ruleName, ruleType, PolicyRuleDeleted, ready); err != nil {
				return err
			}
		}
		return nil
	case *kyverno.Policy:
		for _, rule := range autogen.ComputeRules(inputPolicy) {
			policyValidationMode, err := metrics.ParsePolicyValidationMode(inputPolicy.Spec.GetValidationFailureAction())
			if err != nil {
				return err
			}
			policyBackgroundMode := metrics.ParsePolicyBackgroundMode(inputPolicy)
			policyType := metrics.Namespaced
			policyNamespace := inputPolicy.GetNamespace()
			policyName := inputPolicy.GetName()
			ruleName := rule.Name
			ruleType := metrics.ParseRuleType(rule)
			ready := inputPolicy.IsReady()

			if err = pc.registerPolicyRuleInfoMetric(policyValidationMode, policyType, policyBackgroundMode, policyNamespace, policyName, ruleName, ruleType, PolicyRuleDeleted, ready); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("wrong input type provided %T. Only kyverno.Policy and kyverno.ClusterPolicy allowed", inputPolicy)
	}

}
