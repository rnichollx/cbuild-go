package ccommon

type SuggestedDependency struct {
	URL     string `yaml:"url"`
	Type    string `yaml:"type"`
	Version string `yaml:"version"`
}

type CSetupLists struct {
	Dependencies  []string `yaml:"dependencies"`
	SuggestedDeps []string `yaml:"suggested_dependencies"`
	CxxVersion    string   `yaml:"cxx_version"`
}
