package ccommon

type CSetupLists struct {
	SuggestedSources map[string]CodeSource `yaml:"suggested_dep_sources"`
	DefaultConfig    TargetConfiguration   `yaml:"default_configuration"`
}
