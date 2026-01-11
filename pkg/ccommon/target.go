package ccommon

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"gitlab.com/rpnx/cbuild-go/pkg/cmake"

	"gopkg.in/yaml.v3"
)

type TargetContext struct {
	Name   string
	Config TargetConfiguration
}

type TargetConfiguration struct {
	/// The source to use from workspace sources, if empty, it's the source
	/// of the same name as the target.
	Source string `yaml:"source,omitempty"`

	/// If not empty, a relative path from within the source to treat as the target source when building.
	/// This can be useful for example, if a git repository has multiple subproject targets to build.
	RootPath string `yaml:"root_path,omitempty"`

	/// A list of build targets that this target depends on
	Depends []string `yaml:"depends"`

	/// The project type, such as CMake.
	ProjectType string `yaml:"project_type"`

	/// The CMake moniker used for find_package, if it's different from the target name
	CMakePackageName string `yaml:"cmake_package_name,omitempty"`

	FindPackageRoot        *string `yaml:"find_package_root,omitempty"`
	Staged                 *bool   `yaml:"staged,omitempty"`
	ExternalSourceOverride *string `yaml:"external_source_override,omitempty"`

	/// If config is generated in a subdirectory of the build tree, it needs to be  put into OverrirdeCMakeConfigPath
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

// CMakeConfigureArgs returns the arguments to pass to cmake when configuring the module
func (t *TargetContext) CMakeConfigureArgs(ctx context.Context, workspace *WorkspaceContext, bp TargetBuildParameters) ([]string, error) {

	args := []string{}

	src, err := t.CMakeSourcePath(ctx, workspace)
	if err != nil {
		return nil, err
	}
	src, err = filepath.Abs(src)
	if err != nil {
		return nil, err
	}

	args = append(args, "-S")
	args = append(args, src)

	bld, err := t.CMakeBuildPath(ctx, workspace, bp)
	if err != nil {
		return nil, err
	}
	bld, err = filepath.Abs(bld)
	if err != nil {
		return nil, err
	}

	args = append(args, "-B")
	args = append(args, bld)

	args = append(args, "-G")
	args = append(args, "Ninja")

	args = append(args, fmt.Sprintf("-DCMAKE_BUILD_TYPE=%s", bp.BuildType))

	cxxStandard := ""
	if t.Config.CxxStandard != nil {
		cxxStandard = *t.Config.CxxStandard
	} else if workspace.Config.CXXVersion != "" {
		cxxStandard = workspace.Config.CXXVersion
	}

	if cxxStandard != "" {
		args = append(args, fmt.Sprintf("-DCMAKE_CXX_STANDARD=%s", cxxStandard))
	}

	toolchainFile, err := workspace.ToolchainFilePath(ctx, &t.Config, bp)
	if err != nil {
		return nil, err
	}
	if toolchainFile != "" {
		args = append(args, fmt.Sprintf("-DCMAKE_TOOLCHAIN_FILE=%s", toolchainFile))
	}

	stagedPaths := []string{}
	for _, dep := range t.Config.Depends {
		parts := strings.SplitN(dep, "/", 2)
		targetName := parts[0]

		depMod, err := workspace.GetTarget(ctx, targetName)
		if err != nil {
			return nil, err
		}

		if depMod.Config.Staged != nil && *depMod.Config.Staged {
			stagingPath, err := depMod.CMakeStagingPath(ctx, workspace, bp)
			if err != nil {
				return nil, err
			}
			stagingPath, err = filepath.Abs(stagingPath)
			if err != nil {
				return nil, err
			}
			stagedPaths = append(stagedPaths, stagingPath)
		}
	}

	paths := strings.Join(stagedPaths, ";")
	args = append(args, fmt.Sprintf("-DCMAKE_PREFIX_PATH=%s", paths))
	args = append(args, fmt.Sprintf("-DCMAKE_MODULE_PATH=%s", paths))

	for _, dep := range t.Config.Depends {
		parts := strings.SplitN(dep, "/", 2)
		targetName := parts[0]

		mod, err := workspace.GetTarget(ctx, targetName)
		if err != nil {
			return nil, err
		}

		if mod.Config.Staged != nil && *mod.Config.Staged {
			continue
		}

		mod_args, err := mod.CMakeDependencyArgs(ctx, workspace, bp)
		if err != nil {
			return nil, err
		}
		args = append(args, mod_args...)
	}

	args = append(args, t.Config.ExtraCMakeConfigureArgs...)

	for optName, opt := range t.Config.CMakeOptions {
		if opt.Type != "" {
			args = append(args, fmt.Sprintf("-D%s:%s=%s", optName, opt.Type, opt.Value))
		} else {
			args = append(args, fmt.Sprintf("-D%s=%s", optName, opt.Value))
		}
	}

	return args, nil
}

func (t *TargetContext) CMakeSourcePath(ctx context.Context, workspace *WorkspaceContext) (string, error) {
	if t.Name == "" {
		panic("target context must have a name")
	}

	if t.Config.ExternalSourceOverride != nil {
		path := *t.Config.ExternalSourceOverride
		if filepath.IsAbs(path) {
			return path, nil
		}
		return filepath.Join(workspace.WorkspacePath, "sources", path), nil
	}

	sourceName := t.Config.Source
	if sourceName == "" {
		sourceName = t.Name
	}

	sourcePath := filepath.Join(workspace.WorkspacePath, "sources", sourceName)
	if t.Config.RootPath != "" {
		sourcePath = filepath.Join(sourcePath, t.Config.RootPath)
	}
	return sourcePath, nil
}

func (t *TargetContext) CMakeConfigPath(ctx context.Context, workspace *WorkspaceContext, bp TargetBuildParameters) (string, error) {

	if t.Name == "" {
		panic("target context must have a name")
	}
	buildPath, err := t.CMakeBuildPath(ctx, workspace, bp)
	if err != nil {
		return "", err
	}

	if t.Config.OverrideCMakeConfigPath != nil {
		return filepath.Join(buildPath, *t.Config.OverrideCMakeConfigPath), nil
	}

	return buildPath, nil
}

func (t *TargetContext) CMakeBuildPath(ctx context.Context, workspace *WorkspaceContext, bp TargetBuildParameters) (string, error) {
	return filepath.Join(workspace.WorkspacePath, "buildspaces", bp.Toolchain, t.Name, bp.BuildType), nil
}

func (t *TargetContext) CMakeStagingPath(ctx context.Context, workspace *WorkspaceContext, bp TargetBuildParameters) (string, error) {
	return filepath.Join(workspace.WorkspacePath, "staging", bp.Toolchain, bp.BuildType, t.Name), nil
}

func (t *TargetContext) CMakeExportPath(ctx context.Context, workspace *WorkspaceContext, bp TargetBuildParameters) (string, error) {
	return filepath.Join(workspace.WorkspacePath, "exports", bp.Toolchain, t.Name, bp.BuildType), nil
}

// CMakeDependencyArgs returns the arguments to pass to cmake when configuring another module that depends on this module
func (t *TargetContext) CMakeDependencyArgs(ctx context.Context, workspace *WorkspaceContext, bp TargetBuildParameters) ([]string, error) {
	args := []string{}

	if t.Config.Staged != nil && *t.Config.Staged {
		stagingPath, err := t.CMakeStagingPath(ctx, workspace, bp)
		if err != nil {
			return nil, err
		}
		stagingPath, err = filepath.Abs(stagingPath)
		if err != nil {
			return nil, err
		}
		args = append(args, fmt.Sprintf("-DCMAKE_PREFIX_PATH=%s", stagingPath))
		args = append(args, fmt.Sprintf("-DCMAKE_MODULE_PATH=%s", stagingPath))
		return args, nil
	}

	packageName := t.Config.CMakePackageName
	if packageName == "" {
		packageName = t.Name
	}

	if packageName != "" {
		dirname := packageName + "_DIR"

		configPath, err := t.CMakeConfigPath(ctx, workspace, bp)
		if err != nil {
			return nil, err
		}
		configPath, err = filepath.Abs(configPath)
		if err != nil {
			return nil, err
		}
		args = append(args, fmt.Sprintf("-D%s=%s", dirname, configPath))
	}

	if t.Config.FindPackageRoot != nil {
		dirname := *t.Config.FindPackageRoot + "_ROOT"

		sourcePath, err := t.CMakeSourcePath(ctx, workspace)
		if err != nil {
			return nil, err
		}
		sourcePath, err = filepath.Abs(sourcePath)
		if err != nil {
			return nil, err
		}

		args = append(args, fmt.Sprintf("-D%s=%s", dirname, sourcePath))
	}

	return args, nil
}
