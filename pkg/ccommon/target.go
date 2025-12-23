package ccommon

import (
	"gitlab.com/rpnx/cbuild-go/pkg/cmake"

	"gopkg.in/yaml.v3"
)

type TargetConfiguration struct {
	Depends                 []string                `yaml:"depends"`
	ProjectType             string                  `yaml:"project_type"`
	CMakePackageName        string                  `yaml:"cmake_package_name,omitempty"`
	FindPackageRoot         *string                 `yaml:"find_package_root,omitempty"`
	Staged                  *bool                   `yaml:"staged,omitempty"`
	ExternalSourceOverride  *string                 `yaml:"external_source_override,omitempty"`
	OverrideCMakeConfigPath *string                 `yaml:"override_cmake_config_path,omitempty"`
	ExtraCMakeConfigureArgs []string                `yaml:"extra_cmake_configure_args,omitempty"`
	CMakeOptions            map[string]cmake.Option `yaml:"cmake_options,omitempty"`
	CxxStandard             *string                 `yaml:"cxx_standard,omitempty"`
}

func (m *TargetConfiguration) MarshalYAML() (interface{}, error) {
	type Alias TargetConfiguration
	node := &yaml.Node{}
	err := node.Encode((*Alias)(m))
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == "extra_cmake_configure_args" {
			valueNode := node.Content[i+1]
			valueNode.Style = yaml.FlowStyle
			for _, item := range valueNode.Content {
				item.Style = yaml.DoubleQuotedStyle
			}
			break
		}
	}

	return node, nil
}
