package ccommon

type CSetupLists struct {
	SuggestedSources     map[string]CodeSource  `yaml:"suggested_dep_sources"`
	DefaultConfig        TargetConfiguration    `yaml:"default_configuration"`
	ProjectDefaultConfig *CSetupProjectDefaults `yaml:"project_default_configuration,omitempty"`
}

type CSetupProjectDefaults struct {
	CxxVersion string `yaml:"cxx_version,omitempty"`
}
