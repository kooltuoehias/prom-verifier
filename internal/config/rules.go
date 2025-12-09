package config

// RuleFile matches the structure of a Prometheus rule file.
type RuleFile struct {
	Groups []RuleGroup `yaml:"groups"`
}

// RuleGroup represents a group of rules in the Prometheus configuration.
type RuleGroup struct {
	Name  string `yaml:"name"`
	Rules []Rule `yaml:"rules"`
}

// Rule represents a single alerting rule.
type Rule struct {
	Alert       string            `yaml:"alert"`
	Expr        string            `yaml:"expr"`
	For         string            `yaml:"for"`
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations"`
}
