package ccommon

type SuggestedDepSource struct {
	URL     string `yaml:"url"`
	Type    string `yaml:"type"`
	Version string `yaml:"version"`
}

type CSetupLists struct {
	Dependencies  []string                      `yaml:"dependencies"`
	SuggestedDeps map[string]SuggestedDepSource `yaml:"suggested_dep_sources"`
	CxxVersion    string                        `yaml:"cxx_version"`
}
