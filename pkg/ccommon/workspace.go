package ccommon

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gitlab.com/rpnx/cbuild-go/pkg/cli"
	"gitlab.com/rpnx/cbuild-go/pkg/cmake"
	"gitlab.com/rpnx/cbuild-go/pkg/host"
	"gitlab.com/rpnx/cbuild-go/pkg/system"

	"gopkg.in/yaml.v3"
)

type WorkspaceContext struct {
	Config        WorkspaceConfig
	WorkspacePath string
	DownloadDeps  bool
}

type WorkspaceConfig struct {
	Sources map[string]*CodeSource          `yaml:"sources"`
	Targets map[string]*TargetConfiguration `yaml:"targets"`

	CMakeBinary    *string  `yaml:"cmake_binary"`
	CXXVersion     string   `yaml:"cxx_version"`
	Configurations []string `yaml:"configurations"`
}

func (w *WorkspaceContext) Load(ctx context.Context, path string) error {
	w.WorkspacePath = path
	// Load the configuration from the file
	yamlFile, err := os.ReadFile(filepath.Join(path, "cbuild_workspace.yml"))
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	w.Config = WorkspaceConfig{}

	err = yaml.Unmarshal(yamlFile, &w.Config)

	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	if len(w.Config.Configurations) == 0 {
		w.Config.Configurations = []string{"Debug", "Release"}
	}

	return nil
}

func (w *WorkspaceContext) GetTarget(ctx context.Context, name string) (*TargetContext, error) {
	targetConfig, ok := w.Config.Targets[name]
	if !ok {
		return nil, fmt.Errorf("target %s not found in workspace", name)
	}

	return &TargetContext{
		Name:   name,
		Config: *targetConfig,
	}, nil
}

func (w *WorkspaceContext) Save(ctx context.Context) error {
	yamlFile, err := yaml.Marshal(w.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(filepath.Join(w.WorkspacePath, "cbuild_workspace.yml"), yamlFile, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (w *WorkspaceContext) GenerateToolchainFile(ctx context.Context, opts *CMakeGenerateToolchainFileOptions, systemName system.Platform, systemProcessor system.Processor, targetPath string) error {
	return cmake.GenerateToolchainFile(ctx, cmake.GenerateToolchainFileOptions{
		CCompiler:       opts.CCompiler,
		CXXCompiler:     opts.CXXCompiler,
		Linker:          opts.Linker,
		ExtraCXXFlags:   opts.ExtraCXXFlags,
		SystemPlatform:  systemName,
		SystemProcessor: systemProcessor,
		WorkspaceDir:    w.WorkspacePath,
		OutputFile:      targetPath,
	})
}

func (w *WorkspaceContext) LoadToolchain(ctx context.Context, toolchainName string) (*Toolchain, string, error) {
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

func (w *WorkspaceContext) ToolchainFilePath(ctx context.Context, modConfig *TargetConfiguration, bp TargetBuildParameters) (string, error) {
	tc, tcPath, err := w.LoadToolchain(ctx, bp.Toolchain)
	if err != nil {
		return "", fmt.Errorf("failed to load toolchain: %w", err)
	}

	hostPlatform := fmt.Sprintf("host-%s-%s", host.DetectHostPlatform().StringLower(), host.DetectHostProcessor().StringLower())
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

func (w *WorkspaceContext) Prebuild(ctx context.Context, bp TargetBuildParameters) (string, error) {
	tc, _, err := w.LoadToolchain(ctx, bp.Toolchain)
	if err != nil {
		return "", fmt.Errorf("failed to load toolchain: %w", err)
	}

	hostPlatform := fmt.Sprintf("host-%s-%s", host.DetectHostPlatform().StringLower(), host.DetectHostProcessor().StringLower())
	if tcf, ok := tc.CMakeToolchain[hostPlatform]; ok {
		tcfPath, err := w.ToolchainFilePath(ctx, nil, bp)
		if err != nil {
			return "", err
		}
		if tcf.Generate != nil {
			err := w.GenerateToolchainFile(ctx, tcf.Generate, tc.TargetSystem, tc.TargetArch, tcfPath)
			if err != nil {
				return "", fmt.Errorf("failed to generate toolchain file: %w", err)
			}
		}
		return tcfPath, nil
	}
	return "", nil
}

func (w *WorkspaceContext) Build(ctx context.Context, bp TargetBuildParameters) error {
	_, err := w.Prebuild(ctx, bp)
	if err != nil {
		return err
	}

	var builtModules = make(map[string]bool)

	for name := range w.Config.Targets {
		mod, err := w.GetTarget(ctx, name)
		if err != nil {
			return err
		}
		err = w.buildModule(ctx, mod, name, builtModules, bp)
		if err != nil {
			return fmt.Errorf("failed to build module %s: %w", name, err)
		}
	}

	return nil
}

func (w *WorkspaceContext) BuildTarget(ctx context.Context, targetName string, bp TargetBuildParameters) error {
	_, err := w.Prebuild(ctx, bp)
	if err != nil {
		return err
	}

	var builtModules = make(map[string]bool)

	mod, err := w.GetTarget(ctx, targetName)
	if err != nil {
		return err
	}

	return w.buildModule(ctx, mod, targetName, builtModules, bp)
}

func (w *WorkspaceContext) BuildDependencies(ctx context.Context, targetName string, bp TargetBuildParameters) error {
	_, err := w.Prebuild(ctx, bp)
	if err != nil {
		return err
	}

	var builtModules = make(map[string]bool)

	mod, err := w.GetTarget(ctx, targetName)
	if err != nil {
		return err
	}

	for _, dep := range mod.Config.Depends {
		parts := strings.SplitN(dep, "/", 2)
		depTargetName := parts[0]
		depMod, err := w.GetTarget(ctx, depTargetName)
		if err != nil {
			return err
		}
		err = w.buildModule(ctx, depMod, depTargetName, builtModules, bp)
		if err != nil {
			return fmt.Errorf("failed to build dependency %s: %w", depTargetName, err)
		}
	}

	return nil
}

func (w *WorkspaceContext) CleanTarget(ctx context.Context, targetName string, bp TargetBuildParameters) error {
	mod, err := w.GetTarget(ctx, targetName)
	if err != nil {
		return err
	}

	buildPath, err := mod.CMakeBuildPath(ctx, w, bp)
	if err != nil {
		return err
	}

	if bp.DryRun {
		fmt.Printf("dry-run: would delete build path %s\n", buildPath)
		return nil
	}
	return os.RemoveAll(buildPath)
}

func (w *WorkspaceContext) Exec(ctx context.Context, command string, args []string, dryRun bool) error {
	fmt.Printf("Executing: %s", command)
	for _, arg := range args {
		fmt.Printf(" %s", arg)
	}
	fmt.Println()

	if dryRun {
		return nil
	}

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (w *WorkspaceContext) buildModule(ctx context.Context, mod *TargetContext, modname string, builtModules map[string]bool, bp TargetBuildParameters) error {

	if builtModules[modname] {
		return nil
	}

	for _, dep := range mod.Config.Depends {
		parts := strings.SplitN(dep, "/", 2)
		targetName := parts[0]
		depMod, err := w.GetTarget(ctx, targetName)
		if err != nil {
			return err
		}
		err = w.buildModule(ctx, depMod, targetName, builtModules, bp)
		if err != nil {
			return fmt.Errorf("failed to build dependency %s: %w", targetName, err)
		}
	}

	cmakeBinary := "cmake"
	if w.Config.CMakeBinary != nil {
		cmakeBinary = *w.Config.CMakeBinary
	}

	if mod.Config.ProjectType != "" && !strings.EqualFold(mod.Config.ProjectType, "CMake") {
		return fmt.Errorf("unsupported project type: %s", mod.Config.ProjectType)
	}

	cMakeConfigureArgs, err := mod.CMakeConfigureArgs(ctx, w, bp)
	if err != nil {
		return fmt.Errorf("failed to get cmake configure args: %w", err)
	}

	err = w.Exec(ctx, cmakeBinary, cMakeConfigureArgs, bp.DryRun)
	if err != nil {
		return fmt.Errorf("failed to configure module %s: %w", modname, err)
	}

	buildPath, err := mod.CMakeBuildPath(ctx, w, bp)
	if err != nil {
		return fmt.Errorf("failed to get build path: %w", err)
	}
	buildPath, err = filepath.Abs(buildPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute build path: %w", err)
	}

	// Build the module
	buildCmd := []string{"--build", buildPath, "--config", bp.BuildType}

	err = w.Exec(ctx, cmakeBinary, buildCmd, bp.DryRun)
	if err != nil {
		return fmt.Errorf("failed to build module %s: %w", modname, err)
	}

	if mod.Config.Staged != nil && *mod.Config.Staged {
		stagingPath, err := mod.CMakeStagingPath(ctx, w, bp)
		if err != nil {
			return fmt.Errorf("failed to get staging path: %w", err)
		}
		stagingPath, err = filepath.Abs(stagingPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute staging path: %w", err)
		}

		installCmd := []string{"--install", buildPath, "--prefix", stagingPath, "--config", bp.BuildType}
		err = w.Exec(ctx, cmakeBinary, installCmd, bp.DryRun)
		if err != nil {
			return fmt.Errorf("failed to install module %s to staging: %w", modname, err)
		}
	}

	builtModules[modname] = true

	return nil
}

func (w *WorkspaceContext) ProcessCSetupConfig(ctx context.Context, sourceName string) error {
	sourcePath := filepath.Join(w.WorkspacePath, "sources", sourceName)

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

	// Update targets that use this source
	for targetName, targetConfig := range w.Config.Targets {
		targetSourceName := targetConfig.Source
		if targetSourceName == "" {
			targetSourceName = targetName
		}

		if targetSourceName == sourceName {
			externalOverride := targetConfig.ExternalSourceOverride
			source := targetConfig.Source
			newConfig := csetup.DefaultConfig
			newConfig.ExternalSourceOverride = externalOverride
			newConfig.Source = source
			w.Config.Targets[targetName] = &newConfig
		}
	}

	// Process Suggested Dependencies
	for depName, sdep := range csetup.SuggestedSources {
		if err := sdep.ValidateWeb(); err != nil {
			return fmt.Errorf("invalid suggested source for dependency %s: %w", depName, err)
		}

		if _, exists := w.Config.Targets[depName]; !exists {
			autoDownload := cli.GetBool(ctx, cli.FlagKey(FlagDownload))
			if !autoDownload {
				fmt.Printf("Dependency '%s' is not present in sources, source '%s' suggests getting it from '%s', download it? [Y/n] ", depName, sourceName, sdep.From())
				response, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("error reading input: %w", err)
				}
				response = strings.ToLower(strings.TrimSpace(response))
				if response == "" || response == "y" || response == "yes" {
					autoDownload = true
				}
			}

			if autoDownload {
				if w.Config.Sources == nil {
					w.Config.Sources = make(map[string]*CodeSource)
				}
				w.Config.Sources[depName] = &sdep

				err := w.DownloadSource(ctx, depName)
				if err != nil {
					return fmt.Errorf("failed to download '%s': %w", depName, err)
				}

				// Add to workspace targets
				w.Config.Targets[depName] = &TargetConfiguration{
					Source: depName,
				}
				fmt.Printf("Added target '%s' to workspace.\n", depName)

				// Also add it as a dependency to the targets using current source
				for targetName, targetConfig := range w.Config.Targets {
					targetSourceName := targetConfig.Source
					if targetSourceName == "" {
						targetSourceName = targetName
					}
					if targetSourceName == sourceName {
						found := false
						for _, d := range targetConfig.Depends {
							if d == depName {
								found = true
								break
							}
						}
						if !found {
							targetConfig.Depends = append(targetConfig.Depends, depName)
							fmt.Printf("Added dependency '%s' to target '%s'.\n", depName, targetName)
						}
					}
				}

				// Recursively process the new target's csetup file
				err = w.ProcessCSetupConfig(ctx, depName)
				if err != nil {
					return fmt.Errorf("error processing csetup file for %s: %w", depName, err)
				}
			}
		}
	}

	return w.Save(ctx)
}

func (w *WorkspaceContext) ProcessCSetupFile(ctx context.Context, targetName string) error {
	targetConfig, ok := w.Config.Targets[targetName]
	if !ok {
		return fmt.Errorf("target %s not found in workspace", targetName)
	}

	sourceName := targetConfig.Source
	if sourceName == "" {
		sourceName = targetName
	}

	return w.ProcessCSetupConfig(ctx, sourceName)
}

func (w *WorkspaceContext) DownloadSource(ctx context.Context, sourceName string) error {
	source, ok := w.Config.Sources[sourceName]
	if !ok {
		return fmt.Errorf("source %s not found in workspace configuration", sourceName)
	}

	err := w.Get(ctx, sourceName, *source)
	if err != nil {
		return err
	}

	return w.Save(ctx)
}

func (ws *WorkspaceContext) ListToolchains(ctx context.Context) ([]string, error) {
	toolchainsDir := filepath.Join(ws.WorkspacePath, "toolchains")
	entries, err := os.ReadDir(toolchainsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read toolchains directory: %w", err)
	}

	var toolchains []string
	for _, entry := range entries {
		if entry.IsDir() {
			toolchains = append(toolchains, entry.Name())
		}
	}

	return toolchains, nil
}

func (ws *WorkspaceContext) ListTargets(ctx context.Context) []string {
	var targets []string
	for k, _ := range ws.Config.Targets {
		targets = append(targets, k)
	}
	return targets
}

func (w *WorkspaceContext) Init(ctx context.Context, reinit bool) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(w.WorkspacePath, 0755); err != nil {
		return fmt.Errorf("error creating workspace directory: %w", err)
	}

	workspaceConfig := filepath.Join(w.WorkspacePath, "cbuild_workspace.yml")
	if _, err := os.Stat(workspaceConfig); err == nil {
		if !reinit {
			return fmt.Errorf("%s already exists. Use --reinit to overwrite", workspaceConfig)
		} else {
			// Delete toolchains, sources and buildspaces
			dirsToDelete := []string{"toolchains", "sources", "buildspaces"}
			for _, d := range dirsToDelete {
				dirPath := filepath.Join(w.WorkspacePath, d)
				fmt.Printf("Cleaning %s...\n", dirPath)
				os.RemoveAll(dirPath)
			}
		}
	}

	w.Config = WorkspaceConfig{
		Targets:    make(map[string]*TargetConfiguration),
		CXXVersion: "20",
	}

	return w.Save(ctx)
}

func (w *WorkspaceContext) AddDependency(ctx context.Context, targetName string, depName string) error {
	target, ok := w.Config.Targets[targetName]
	if !ok {
		return fmt.Errorf("target %s not found in workspace", targetName)
	}

	// Check if dependency already exists
	for _, d := range target.Depends {
		if d == depName {
			fmt.Printf("Dependency %s already exists for %s\n", depName, targetName)
			return nil
		}
	}

	target.Depends = append(target.Depends, depName)
	return w.Save(ctx)
}

func (w *WorkspaceContext) RemoveDependency(ctx context.Context, targetName string, depName string) error {
	target, ok := w.Config.Targets[targetName]
	if !ok {
		return fmt.Errorf("target %s not found in workspace", targetName)
	}

	newDepends := []string{}
	found := false
	for _, d := range target.Depends {
		if d == depName {
			found = true
			continue
		}
		newDepends = append(newDepends, d)
	}

	if !found {
		fmt.Printf("Dependency %s not found for %s\n", depName, targetName)
		return nil
	}

	target.Depends = newDepends
	return w.Save(ctx)
}

func (w *WorkspaceContext) RemoveSource(ctx context.Context, sourceName string, removeFolder bool) error {
	if w.Config.Sources == nil {
		return fmt.Errorf("no sources defined in workspace")
	}

	if _, ok := w.Config.Sources[sourceName]; !ok {
		return fmt.Errorf("source %s not found in workspace", sourceName)
	}

	delete(w.Config.Sources, sourceName)

	err := w.Save(ctx)
	if err != nil {
		return fmt.Errorf("error saving workspace: %w", err)
	}

	fmt.Printf("Removed source %s from workspace\n", sourceName)

	if removeFolder {
		sourceDir := filepath.Join(w.WorkspacePath, "sources", sourceName)
		if _, err := os.Stat(sourceDir); err == nil {
			fmt.Printf("Deleting source folder: %s\n", sourceDir)
			err = os.RemoveAll(sourceDir)
			if err != nil {
				return fmt.Errorf("error deleting source folder: %w", err)
			}
		} else {
			fmt.Printf("Source folder %s not found, skipping deletion.\n", sourceDir)
		}
	} else {
		fmt.Printf("Note: files in sources/%s were NOT deleted. Use -X to delete them.\n", sourceName)
	}
	return nil
}

func (w *WorkspaceContext) RemoveTarget(ctx context.Context, targetName string) error {
	if _, ok := w.Config.Targets[targetName]; !ok {
		return fmt.Errorf("target %s not found in workspace", targetName)
	}

	delete(w.Config.Targets, targetName)

	err := w.Save(ctx)
	if err != nil {
		return fmt.Errorf("error saving workspace: %w", err)
	}

	fmt.Printf("Removed target %s from workspace\n", targetName)
	return nil
}

func (w *WorkspaceContext) RemoveProject(ctx context.Context, sourceName string, removeFolder bool) error {
	sourceFound := false
	if w.Config.Sources != nil {
		if _, ok := w.Config.Sources[sourceName]; ok {
			delete(w.Config.Sources, sourceName)
			sourceFound = true
		}
	}

	targetsToRemove := []string{}
	for targetName, targetConfig := range w.Config.Targets {
		targetSource := targetConfig.Source
		if targetSource == "" {
			targetSource = targetName
		}
		if targetSource == sourceName {
			targetsToRemove = append(targetsToRemove, targetName)
		}
	}

	for _, targetName := range targetsToRemove {
		delete(w.Config.Targets, targetName)
	}

	if !sourceFound && len(targetsToRemove) == 0 {
		return fmt.Errorf("source or targets for %s not found in workspace", sourceName)
	}

	err := w.Save(ctx)
	if err != nil {
		return fmt.Errorf("error saving workspace: %w", err)
	}

	if sourceFound {
		fmt.Printf("Removed source %s from workspace\n", sourceName)
	}
	for _, targetName := range targetsToRemove {
		fmt.Printf("Removed target %s from workspace\n", targetName)
	}

	if removeFolder {
		sourceDir := filepath.Join(w.WorkspacePath, "sources", sourceName)
		if _, err := os.Stat(sourceDir); err == nil {
			fmt.Printf("Deleting source folder: %s\n", sourceDir)
			err = os.RemoveAll(sourceDir)
			if err != nil {
				return fmt.Errorf("error deleting source folder: %w", err)
			}
		} else {
			fmt.Printf("Source folder %s not found, skipping deletion.\n", sourceDir)
		}
	} else if sourceFound {
		fmt.Printf("Note: files in sources/%s were NOT deleted. Use -X to delete them.\n", sourceName)
	}

	return nil
}

func (w *WorkspaceContext) SetCXXVersion(ctx context.Context, version string, source string) error {
	if source != "" {
		target, ok := w.Config.Targets[source]
		if !ok {
			return fmt.Errorf("source %s not found in workspace", source)
		}
		target.CxxStandard = &version
		fmt.Printf("Set CXX version for %s to %s\n", source, version)
	} else {
		w.Config.CXXVersion = version
		fmt.Printf("Set global CXX version to %s\n", version)
	}

	return w.Save(ctx)
}

func (w *WorkspaceContext) SetStaging(ctx context.Context, source string, enabled bool) error {
	target, ok := w.Config.Targets[source]
	if !ok {
		return fmt.Errorf("source %s not found in workspace", source)
	}

	target.Staged = &enabled

	err := w.Save(ctx)
	if err != nil {
		return err
	}

	if enabled {
		fmt.Printf("Enabled staging for %s\n", source)
	} else {
		fmt.Printf("Disabled staging for %s\n", source)
	}
	return nil
}

func (w *WorkspaceContext) AddConfiguration(ctx context.Context, configName string) error {
	for _, cfg := range w.Config.Configurations {
		if cfg == configName {
			fmt.Printf("Configuration %s already exists\n", configName)
			return nil
		}
	}

	w.Config.Configurations = append(w.Config.Configurations, configName)
	err := w.Save(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Added configuration %s\n", configName)
	return nil
}

func (w *WorkspaceContext) RemoveConfiguration(ctx context.Context, configName string) error {
	found := false
	newConfigs := []string{}
	for _, cfg := range w.Config.Configurations {
		if cfg == configName {
			found = true
			continue
		}
		newConfigs = append(newConfigs, cfg)
	}

	if !found {
		return fmt.Errorf("configuration %s not found", configName)
	}

	w.Config.Configurations = newConfigs
	err := w.Save(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Removed configuration %s\n", configName)
	return nil
}

func (ws *WorkspaceContext) DetectToolchains(ctx context.Context) error {
	hostOS := host.DetectHostPlatform()
	hostProcessor := host.DetectHostProcessor()
	hostKey := fmt.Sprintf("host-%s-%s", hostOS.StringLower(), hostProcessor.StringLower())

	targetSystem := hostOS
	targetArch := hostProcessor

	detectors := []struct {
		name          string
		cCompiler     string
		cxxCompiler   string
		extraCXXFlags []string
	}{
		{"system-gcc", "gcc", "g++", nil},
		{"system-clang", "clang", "clang++", nil},
		{"system-clang-libcxx", "clang", "clang++", []string{"-stdlib=libc++"}},
		{"system-gcc-libcxx", "gcc", "g++", []string{"-stdlib=libc++"}},
	}

	toolchainsDir := filepath.Join(ws.WorkspacePath, "toolchains")
	err := os.MkdirAll(toolchainsDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create toolchains directory: %w", err)
	}

	for _, d := range detectors {
		_, errC := exec.LookPath(d.cCompiler)
		_, errCXX := exec.LookPath(d.cxxCompiler)

		if errC == nil && errCXX == nil {

			if hostOS == system.PlatformMac && strings.Contains(d.name, "libcxx") {
				continue
			}

			if d.cCompiler == "gcc" {
				isGCCReal, err := GCCIsRealGCC(d.cCompiler)
				if err != nil {
					return fmt.Errorf("error checking GCCIsRealGCC: %w", err)
				}
				if !isGCCReal {
					continue
				}
			}
			// Check if compilers actually work by building a minimal CMake project
			testDir, err := os.MkdirTemp("", "csetup_detect_test")
			if err != nil {
				return fmt.Errorf("failed to create temp dir: %w", err)
			}
			defer os.RemoveAll(testDir)

			err = os.MkdirAll(filepath.Join(testDir, "build"), 0755)
			if err != nil {
				return fmt.Errorf("failed to create build dir: %w", err)
			}

			err = os.WriteFile(filepath.Join(testDir, "CMakeLists.txt"), []byte("cmake_minimum_required(VERSION 3.10)\nproject(test)\nadd_executable(test main.cpp)\n"), 0644)
			if err != nil {
				return fmt.Errorf("failed to create CMakeLists.txt: %w", err)
			}

			err = os.WriteFile(filepath.Join(testDir, "main.cpp"), []byte("int main() { return 0; }\n"), 0644)
			if err != nil {
				return fmt.Errorf("failed to create main.cpp: %w", err)
			}

			// Generate a temporary toolchain file for the test
			tcFilePath := filepath.Join(testDir, "toolchain.cmake")
			tc := Toolchain{
				TargetArch:   targetArch,
				TargetSystem: targetSystem,
				CMakeToolchain: map[string]CMakeToolchainOptions{
					hostKey: {
						Generate: &CMakeGenerateToolchainFileOptions{
							CCompiler:     d.cCompiler,
							CXXCompiler:   d.cxxCompiler,
							ExtraCXXFlags: d.extraCXXFlags,
						},
					},
				},
			}

			// We need a workspace to call GenerateToolchainFile, but we can call cmake.GenerateToolchainFile directly
			err = ws.GenerateToolchainFile(ctx, tc.CMakeToolchain[hostKey].Generate, targetSystem, targetArch, tcFilePath)
			if err != nil {
				return fmt.Errorf("failed to generate test toolchain file: %w", err)
			}

			// Run CMake configure
			cmd := exec.CommandContext(ctx, "cmake", "-S", testDir, "-B", filepath.Join(testDir, "build"), "-G", "Ninja", "-DCMAKE_TOOLCHAIN_FILE="+tcFilePath)
			err = cmd.Run()
			if err != nil {
				fmt.Printf("Detected %s, but %s cannot build a hello world program, skipping.\n", d.cxxCompiler, d.name)
				continue
			}

			// Run CMake build
			cmd = exec.CommandContext(ctx, "cmake", "--build", filepath.Join(testDir, "build"))
			err = cmd.Run()
			if err != nil {
				fmt.Printf("Detected %s, but %s cannot build a hello world program, skipping.\n", d.cxxCompiler, d.name)
				continue
			}

			fmt.Printf("Detected %s, creating toolchain...\n", d.name)

			finalTc := Toolchain{
				TargetArch:   targetArch,
				TargetSystem: targetSystem,
				CMakeToolchain: map[string]CMakeToolchainOptions{
					hostKey: {
						Generate: &CMakeGenerateToolchainFileOptions{
							CCompiler:     d.cCompiler,
							CXXCompiler:   d.cxxCompiler,
							ExtraCXXFlags: d.extraCXXFlags,
						},
					},
				},
			}

			tcDir := filepath.Join(toolchainsDir, d.name)
			err = os.MkdirAll(tcDir, 0755)
			if err != nil {
				return fmt.Errorf("failed to create toolchain directory for %s: %w", d.name, err)
			}

			yamlFile, err := yaml.Marshal(finalTc)
			if err != nil {
				return fmt.Errorf("failed to marshal toolchain %s: %w", d.name, err)
			}

			err = os.WriteFile(filepath.Join(tcDir, "toolchain.yml"), yamlFile, 0644)
			if err != nil {
				return fmt.Errorf("failed to write toolchain file for %s: %w", d.name, err)
			}
		} else {
			fmt.Printf("Compilers for %s not found (tried %s and %s), skipping.\n", d.name, d.cCompiler, d.cxxCompiler)
		}
	}

	return nil
}

func (ws *WorkspaceContext) GetBuildArgs(ctx context.Context, targetName string, bp TargetBuildParameters) ([]string, error) {
	target, err := ws.GetTarget(ctx, targetName)
	if err != nil {
		return nil, err
	}

	fullArgs, err := target.CMakeConfigureArgs(ctx, ws, bp)
	if err != nil {
		return nil, fmt.Errorf("error getting cmake args: %w", err)
	}

	filteredArgs := []string{}
	for i := 0; i < len(fullArgs); i++ {
		arg := fullArgs[i]
		if arg == "-S" || arg == "-B" {
			i++ // skip the next argument too (the path)
			continue
		}
		filteredArgs = append(filteredArgs, arg)
	}

	return filteredArgs, nil
}

func (w *WorkspaceContext) LoadDefaults(ctx context.Context, sourceName string) error {
	if w.Config.Sources == nil {
		return fmt.Errorf("no sources defined in workspace")
	}

	if _, ok := w.Config.Sources[sourceName]; !ok {
		return fmt.Errorf("source %s not found in workspace", sourceName)
	}

	// Create a new target with that source name if it doesn't exist
	if w.Config.Targets == nil {
		w.Config.Targets = make(map[string]*TargetConfiguration)
	}

	if _, ok := w.Config.Targets[sourceName]; !ok {
		w.Config.Targets[sourceName] = &TargetConfiguration{
			Source: sourceName,
		}
		fmt.Printf("Created target %s for source %s\n", sourceName, sourceName)
	}

	return w.ProcessCSetupConfig(ctx, sourceName)
}

func (w *WorkspaceContext) DropSourceFiles(ctx context.Context, sourceName string) error {
	if _, ok := w.Config.Sources[sourceName]; !ok {
		return fmt.Errorf("source %s not found in workspace configuration", sourceName)
	}

	sourceDir := filepath.Join(w.WorkspacePath, "sources", sourceName)
	if info, err := os.Stat(sourceDir); err == nil && info.IsDir() {
		fmt.Printf("Deleting source folder: %s\n", sourceDir)
		err = os.RemoveAll(sourceDir)
		if err != nil {
			return fmt.Errorf("error deleting source folder %s: %w", sourceDir, err)
		}
	} else {
		// If it's not there, it's already "dropped"
		return nil
	}
	return nil
}
