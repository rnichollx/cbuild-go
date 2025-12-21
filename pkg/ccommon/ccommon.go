package ccommon

import (
	"bufio"
	"cbuild-go/pkg/cmake"
	"cbuild-go/pkg/host"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type CMakeToolchainOptions struct {
	CMakeToolchainFile string                             `yaml:"cmake_toolchain_file,omitempty"`
	Generate           *CMakeGenerateToolchainFileOptions `yaml:"generate,omitempty"`
}

type CMakeGenerateToolchainFileOptions struct {
	CCompiler     string   `yaml:"c_compiler"`
	CXXCompiler   string   `yaml:"cxx_compiler"`
	Linker        string   `yaml:"linker,omitempty"`
	ExtraCXXFlags []string `yaml:"extra_cxx_flags,omitempty"`
}

type Toolchain struct {
	CMakeToolchain map[string]CMakeToolchainOptions `yaml:"cmake_toolchain"`
	TargetArch     string                           `yaml:"target_arch"`
	TargetSystem   string                           `yaml:"target_system"`
}

type BuildParameters struct {
	Toolchain string
	BuildType string
	DryRun    bool
}

type Workspace struct {
	WorkspacePath string `yaml:"-"`

	Targets map[string]*Target `yaml:"targets"`

	CMakeBinary *string `yaml:"cmake_binary"`
	CXXVersion  string  `yaml:"cxx_version"`
}

type CMakeOption struct {
	Type  string `yaml:"type"`
	Value string `yaml:"value"`
}

func (o *CMakeOption) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		o.Value = value.Value
		o.Type = ""
		return nil
	}
	type Alias CMakeOption
	var aux Alias
	if err := value.Decode(&aux); err != nil {
		return err
	}
	o.Type = aux.Type
	o.Value = aux.Value
	return nil
}

func (o CMakeOption) MarshalYAML() (interface{}, error) {
	if o.Type == "" {
		return o.Value, nil
	}
	type Alias CMakeOption
	return Alias(o), nil
}

type Target struct {
	Depends                 []string               `yaml:"depends"`
	ProjectType             string                 `yaml:"project_type"`
	CMakePackageName        string                 `yaml:"cmake_package_name,omitempty"`
	FindPackageRoot         *string                `yaml:"find_package_root,omitempty"`
	Staged                  *bool                  `yaml:"staged,omitempty"`
	ExternalSourceOverride  *string                `yaml:"external_source_override,omitempty"`
	OverrideCMakeConfigPath *string                `yaml:"override_cmake_config_path,omitempty"`
	ExtraCMakeConfigureArgs []string               `yaml:"extra_cmake_configure_args,omitempty"`
	CMakeOptions            map[string]CMakeOption `yaml:"cmake_options,omitempty"`
	CxxStandard             *string                `yaml:"cxx_standard,omitempty"`
}

func (m *Target) MarshalYAML() (interface{}, error) {
	type Alias Target
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
func (m *Target) CMakeConfigureArgs(workspace *Workspace, modname string, bp BuildParameters) ([]string, error) {

	args := []string{}

	src, err := m.CMakeSourcePath(workspace, modname)
	if err != nil {
		return nil, err
	}
	src, err = filepath.Abs(src)
	if err != nil {
		return nil, err
	}

	args = append(args, "-S")
	args = append(args, src)

	bld, err := m.CMakeBuildPath(workspace, modname, bp)
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
	if m.CxxStandard != nil {
		cxxStandard = *m.CxxStandard
	} else if workspace.CXXVersion != "" {
		cxxStandard = workspace.CXXVersion
	}

	if cxxStandard != "" {
		args = append(args, fmt.Sprintf("-DCMAKE_CXX_STANDARD=%s", cxxStandard))
	}

	toolchainFile, err := workspace.ToolchainFilePath(m, bp)
	if err != nil {
		return nil, err
	}
	if toolchainFile != "" {
		args = append(args, fmt.Sprintf("-DCMAKE_TOOLCHAIN_FILE=%s", toolchainFile))
	}

	stagedPaths := []string{}
	for _, dep := range m.Depends {
		parts := strings.SplitN(dep, "/", 2)
		targetName := parts[0]

		depMod, ok := workspace.Targets[targetName]
		if !ok {
			return nil, fmt.Errorf("unknown target %s", targetName)
		}

		if depMod.Staged != nil && *depMod.Staged {
			stagingPath, err := depMod.CMakeStagingPath(workspace, targetName, bp)
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

	if len(stagedPaths) > 0 {
		paths := strings.Join(stagedPaths, ";")
		args = append(args, fmt.Sprintf("-DCMAKE_PREFIX_PATH=%s", paths))
		args = append(args, fmt.Sprintf("-DCMAKE_MODULE_PATH=%s", paths))
	}

	for _, dep := range m.Depends {
		parts := strings.SplitN(dep, "/", 2)
		targetName := parts[0]

		mod, ok := workspace.Targets[targetName]
		if !ok {
			return nil, fmt.Errorf("depend on unknown target %s", targetName)
		}

		if mod.Staged != nil && *mod.Staged {
			continue
		}

		mod_args, err := mod.CMakeDependencyArgs(workspace, targetName, bp)
		if err != nil {
			return nil, err
		}
		args = append(args, mod_args...)
	}

	args = append(args, m.ExtraCMakeConfigureArgs...)

	for optName, opt := range m.CMakeOptions {
		if opt.Type != "" {
			args = append(args, fmt.Sprintf("-D%s:%s=%s", optName, opt.Type, opt.Value))
		} else {
			args = append(args, fmt.Sprintf("-D%s=%s", optName, opt.Value))
		}
	}

	return args, nil
}

func (m *Target) CMakeSourcePath(workspace *Workspace, defaultRoot string) (string, error) {
	path := m.ExternalSourceOverride
	if path == nil {
		return filepath.Join(workspace.WorkspacePath, "sources", defaultRoot), nil
	}

	if filepath.IsAbs(*path) {
		return *path, nil
	}

	return filepath.Join(workspace.WorkspacePath, "sources", *path), nil
}

func (m *Target) CMakeConfigPath(workspace *Workspace, defaultRoot string, bp BuildParameters) (string, error) {

	buildPath, err := m.CMakeBuildPath(workspace, defaultRoot, bp)
	if err != nil {
		return "", err
	}

	if m.OverrideCMakeConfigPath != nil {
		return filepath.Join(buildPath, *m.OverrideCMakeConfigPath), nil
	}

	return buildPath, nil
}

func (m *Target) CMakeBuildPath(workspace *Workspace, defaultRoot string, bp BuildParameters) (string, error) {
	return filepath.Join(workspace.WorkspacePath, "buildspaces", bp.Toolchain, defaultRoot, bp.BuildType), nil
}

func (m *Target) CMakeStagingPath(workspace *Workspace, defaultRoot string, bp BuildParameters) (string, error) {
	return filepath.Join(workspace.WorkspacePath, "staging", bp.Toolchain, bp.BuildType, defaultRoot), nil
}

func (m *Target) CMakeExportPath(workspace *Workspace, defaultRoot string, bp BuildParameters) (string, error) {
	return filepath.Join(workspace.WorkspacePath, "exports", bp.Toolchain, defaultRoot, bp.BuildType), nil
}

// CMakeDependencyArgs returns the arguments to pass to cmake when configuring another module that depends on this module
func (m *Target) CMakeDependencyArgs(workspace *Workspace, modname string, bp BuildParameters) ([]string, error) {
	args := []string{}

	if m.Staged != nil && *m.Staged {
		stagingPath, err := m.CMakeStagingPath(workspace, modname, bp)
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

	packageName := m.CMakePackageName
	if packageName == "" {
		packageName = modname
	}

	if packageName != "" {
		dirname := packageName + "_DIR"

		configPath, err := m.CMakeConfigPath(workspace, modname, bp)
		if err != nil {
			return nil, err
		}
		configPath, err = filepath.Abs(configPath)
		if err != nil {
			return nil, err
		}
		args = append(args, fmt.Sprintf("-D%s=%s", dirname, configPath))
	}

	if m.FindPackageRoot != nil {
		dirname := *m.FindPackageRoot + "_ROOT"

		sourcePath, err := m.CMakeSourcePath(workspace, modname)
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

func (w *Workspace) Load(path string) error {
	w.WorkspacePath = path
	// Load the configuration from the file
	yamlFile, err := os.ReadFile(filepath.Join(path, "cbuild_workspace.yml"))
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(yamlFile, w)

	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

func (w *Workspace) Save() error {
	yamlFile, err := yaml.Marshal(w)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(filepath.Join(w.WorkspacePath, "cbuild_workspace.yml"), yamlFile, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (w *Workspace) GenerateToolchainFile(opts *CMakeGenerateToolchainFileOptions, systemName string, systemProcessor string, targetPath string) error {
	return cmake.GenerateToolchainFile(cmake.GenerateToolchainFileOptions{
		CCompiler:       opts.CCompiler,
		CXXCompiler:     opts.CXXCompiler,
		Linker:          opts.Linker,
		ExtraCXXFlags:   opts.ExtraCXXFlags,
		SystemName:      systemName,
		SystemProcessor: systemProcessor,
		WorkspaceDir:    w.WorkspacePath,
		OutputFile:      targetPath,
	})
}

func (w *Workspace) LoadToolchain(toolchainName string) (*Toolchain, string, error) {
	toolchainDir := filepath.Join(w.WorkspacePath, "toolchains", toolchainName)
	toolchainFile := filepath.Join(toolchainDir, "toolchain.yml")

	yamlFile, err := ioutil.ReadFile(toolchainFile)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read toolchain file: %w", err)
	}

	tc := &Toolchain{}
	err = yaml.Unmarshal(yamlFile, tc)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse toolchain file: %w", err)
	}

	return tc, toolchainDir, nil
}

func (w *Workspace) ToolchainFilePath(mod *Target, bp BuildParameters) (string, error) {
	tc, tcPath, err := w.LoadToolchain(bp.Toolchain)
	if err != nil {
		return "", fmt.Errorf("failed to load toolchain: %w", err)
	}

	hostPlatform := fmt.Sprintf("host-%s-%s", host.DetectHostPlatform(), host.DetectHostArch())
	if tcf, ok := tc.CMakeToolchain[hostPlatform]; ok {
		var tcfPath string
		if tcf.Generate != nil {
			tcfPath = filepath.Join(w.WorkspacePath, "buildspaces", bp.Toolchain, "generated_toolchain.cmake")
		} else {
			tcfPath = filepath.Join(tcPath, tcf.CMakeToolchainFile)
		}

		absTcfPath, err := filepath.Abs(tcfPath)
		if err != nil {
			return "", err
		}
		tcfPath = absTcfPath
		return tcfPath, nil
	}
	return "", nil
}

func (w *Workspace) Prebuild(bp BuildParameters) (string, error) {
	tc, _, err := w.LoadToolchain(bp.Toolchain)
	if err != nil {
		return "", fmt.Errorf("failed to load toolchain: %w", err)
	}

	hostPlatform := fmt.Sprintf("host-%s-%s", host.DetectHostPlatform(), host.DetectHostArch())
	if tcf, ok := tc.CMakeToolchain[hostPlatform]; ok {
		tcfPath, err := w.ToolchainFilePath(nil, bp)
		if err != nil {
			return "", err
		}
		if tcf.Generate != nil {
			err := w.GenerateToolchainFile(tcf.Generate, tc.TargetSystem, tc.TargetArch, tcfPath)
			if err != nil {
				return "", fmt.Errorf("failed to generate toolchain file: %w", err)
			}
		}
		return tcfPath, nil
	}
	return "", nil
}

func (w *Workspace) Build(bp BuildParameters) error {
	_, err := w.Prebuild(bp)
	if err != nil {
		return err
	}

	var builtModules = make(map[string]bool)

	for name, mod := range w.Targets {
		err := w.buildModule(mod, name, builtModules, bp)
		if err != nil {
			return fmt.Errorf("failed to build module %s: %w", name, err)
		}
	}

	return nil
}

func (w *Workspace) BuildTarget(targetName string, bp BuildParameters) error {
	_, err := w.Prebuild(bp)
	if err != nil {
		return err
	}

	var builtModules = make(map[string]bool)

	mod, ok := w.Targets[targetName]
	if !ok {
		return fmt.Errorf("unknown target %s", targetName)
	}

	return w.buildModule(mod, targetName, builtModules, bp)
}

func (w *Workspace) Clean(toolchain string, dryRun bool) error {
	cleanPath := filepath.Join(w.WorkspacePath, "buildspaces")
	if toolchain != "all" && toolchain != "" {
		cleanPath = filepath.Join(cleanPath, toolchain)
	}

	fmt.Printf("Cleaning: %s\n", cleanPath)

	if dryRun {
		return nil
	}

	err := os.RemoveAll(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to clean: %w", err)
	}

	return nil
}

func (w *Workspace) Exec(command string, args []string, dryRun bool) error {
	fmt.Printf("Executing: %s", command)
	for _, arg := range args {
		fmt.Printf(" %s", arg)
	}
	fmt.Println()

	if dryRun {
		return nil
	}

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (w *Workspace) buildModule(mod *Target, modname string, builtModules map[string]bool, bp BuildParameters) error {

	if builtModules[modname] {
		return nil
	}

	// Check if mod has dependencies that need to build first
	for _, dep := range mod.Depends {
		parts := strings.SplitN(dep, "/", 2)
		targetName := parts[0]
		depMod, ok := w.Targets[targetName]
		if !ok {
			return fmt.Errorf("depends on unknown target %s", targetName)
		}
		err := w.buildModule(depMod, targetName, builtModules, bp)
		if err != nil {
			return fmt.Errorf("failed to build dependency %s: %w", targetName, err)
		}
	}

	cmakeBinary := "cmake"
	if w.CMakeBinary != nil {
		cmakeBinary = *w.CMakeBinary
	}

	if mod.ProjectType != "" && !strings.EqualFold(mod.ProjectType, "CMake") {
		return fmt.Errorf("unsupported project type: %s", mod.ProjectType)
	}

	cMakeConfigureArgs, err := mod.CMakeConfigureArgs(w, modname, bp)
	if err != nil {
		return fmt.Errorf("failed to get cmake configure args: %w", err)
	}

	err = w.Exec(cmakeBinary, cMakeConfigureArgs, bp.DryRun)
	if err != nil {
		return fmt.Errorf("failed to configure module %s: %w", modname, err)
	}

	buildPath, err := mod.CMakeBuildPath(w, modname, bp)
	if err != nil {
		return fmt.Errorf("failed to get build path: %w", err)
	}
	buildPath, err = filepath.Abs(buildPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute build path: %w", err)
	}

	// Build the module
	buildCmd := []string{"--build", buildPath, "-j"}

	err = w.Exec(cmakeBinary, buildCmd, bp.DryRun)
	if err != nil {
		return fmt.Errorf("failed to build module %s: %w", modname, err)
	}

	if mod.Staged != nil && *mod.Staged {
		stagingPath, err := mod.CMakeStagingPath(w, modname, bp)
		if err != nil {
			return fmt.Errorf("failed to get staging path: %w", err)
		}
		stagingPath, err = filepath.Abs(stagingPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute staging path: %w", err)
		}

		installCmd := []string{"--install", buildPath, "--prefix", stagingPath}
		err = w.Exec(cmakeBinary, installCmd, bp.DryRun)
		if err != nil {
			return fmt.Errorf("failed to install module %s to staging: %w", modname, err)
		}
	}

	builtModules[modname] = true

	return nil
}

func (w *Workspace) ProcessCSetupFile(targetName string) error {
	target, ok := w.Targets[targetName]
	if !ok {
		return fmt.Errorf("target %s not found in workspace", targetName)
	}

	sourcePath, err := target.CMakeSourcePath(w, targetName)
	if err != nil {
		return fmt.Errorf("failed to get source path for target %s: %w", targetName, err)
	}

	csetupFiles := []string{"csetup.yml", "csetuplists.yml", "CSetup.yml", "CSetupLists.yml"}
	var csetupFile string
	for _, f := range csetupFiles {
		path := filepath.Join(sourcePath, f)
		if _, err := os.Stat(path); err == nil {
			csetupFile = path
			break
		}
	}

	if csetupFile == "" {
		return nil // No csetup file to process
	}

	data, err := os.ReadFile(csetupFile)
	if err != nil {
		return fmt.Errorf("failed to read csetup file %s: %w", csetupFile, err)
	}

	var csetup CSetupLists
	err = yaml.Unmarshal(data, &csetup)
	if err != nil {
		return fmt.Errorf("failed to parse csetup file %s: %w", csetupFile, err)
	}

	reader := bufio.NewReader(os.Stdin)

	// Process Dependencies
	for _, dep := range csetup.Dependencies {
		parts := strings.SplitN(dep, "/", 2)
		depTargetName := parts[0]

		found := false
		for _, existingDep := range target.Depends {
			if existingDep == dep {
				found = true
				break
			}
		}

		if !found {
			target.Depends = append(target.Depends, dep)
			fmt.Printf("Added dependency '%s' to target '%s'.\n", dep, targetName)

			// Also check if the dependency exists in the workspace
			if _, exists := w.Targets[depTargetName]; !exists {
				fmt.Printf("Warning: Dependency '%s' is not defined in the workspace.\n", depTargetName)
			}
		}
	}

	// Process CMakePackageName
	if csetup.CMakePackageName != "" {
		target.CMakePackageName = csetup.CMakePackageName
		fmt.Printf("Set CMake package name for target '%s' to '%s'.\n", targetName, csetup.CMakePackageName)
	}

	// Process Suggested Dependencies
	for depName, sdep := range csetup.SuggestedDeps {
		if _, exists := w.Targets[depName]; !exists {
			fmt.Printf("Dependency '%s' is not present in sources, module '%s' suggests getting it from '%s', download it? [Y/n] ", depName, targetName, sdep.URL)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("error reading input: %w", err)
			}
			response = strings.ToLower(strings.TrimSpace(response))
			if response == "" || response == "y" || response == "yes" {
				fmt.Printf("Downloading '%s' from '%s'...\n", depName, sdep.URL)

				destDir := filepath.Join(w.WorkspacePath, "sources", depName)
				cmd := exec.Command("git", "clone", sdep.URL, destDir)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err := cmd.Run()
				if err != nil {
					return fmt.Errorf("failed to download '%s': %w", depName, err)
				} else {
					// Add to workspace targets
					w.Targets[depName] = &Target{
						ProjectType: "CMake",
					}
					fmt.Printf("Added target '%s' to workspace.\n", depName)

					// Also add it as a dependency to the current target
					target.Depends = append(target.Depends, depName)
					fmt.Printf("Added dependency '%s' to target '%s'.\n", depName, targetName)

					// Recursively process the new target's csetup file
					err = w.ProcessCSetupFile(depName)
					if err != nil {
						return fmt.Errorf("error processing csetup file for %s: %w", depName, err)
					}
				}
			}
		}
	}

	return w.Save()
}
