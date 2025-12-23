# cbuild & csetup

`cbuild` is a tool for building distributions by managing multiple projects simultaneously within a workspace. `csetup` is a companion tool used to initialize and manage these workspaces.

## Workspace Structure

A `cbuild` workspace typically follows this structure:

```
workspace/
  cbuild_workspace.yml        # Main workspace configuration
  toolchains/                 # Toolchain definitions
    <toolchain_name>/
      toolchain.yml
      <toolchain_file>.cmake
  sources/                    # Source code for projects
    <sourcename>/
  buildspaces/                # Build artifacts (temporary)
    <sourcename>/<toolchain>/<config>/
  exports/                    # Build outputs
    <sourcename>/<toolchain>/<config>/
```

## cbuild

`cbuild` is the main build tool. It uses the information in `cbuild_workspace.yml` to build projects.

### Commands

- **`build`** (default): Build the project(s).
- **`clean`**: Remove build artifacts from `buildspaces`.
- **`build-deps <sourcename>`**: Build only the dependencies for a specific source.

### Global Flags

- `-w, --workspace <path>`: Path to the workspace directory (default: current directory or nearest parent with `cbuild_workspace.yml`).
- `-d, --dry-run`: Show commands without executing them.
- `-h, --help`: Show help message.

### Command-specific Flags

- `-c, --config <configs>`: Build configurations to use (e.g., `Debug,Release`), comma-separated.
- `-T, --toolchain <toolchain>`: Specific toolchain to use (default: `all`).
- `-t, --target <target>`: Specific target to build.

## csetup

`csetup` is used for managing the workspace, including adding/removing sources and dependencies.

### Commands

- **`init [--reinit]`**: Initialize a new workspace or reinitialize an existing one.
- **`git-clone <repo_url> <dest_name> [--download-deps]`**: Clone a git repository into the `sources` directory.
- **`add-dependency <sourcename> <dependency>`**: Add a dependency to a source.
- **`remove-dependency <sourcename> <dependency>`**: Remove a dependency from a source.
- **`remove-source <sourcename> [-X, --delete]`**: Remove a source from the workspace.
- **`set-cxx-version [sourcename] <version>`**: Set the C++ version for a source or the whole workspace.
- **`enable-staging <sourcename>`**: Enable staging for a source. Staged targets are built against the installed 
         outputs, instead of a build tree.
- **`disable-staging <sourcename>`**: Disable staging for a source.
- **`list-sources`**: List all sources in the workspace.
- **`get-args <sourcename>`**: Get the build arguments that would be passed to the build system (e.g., CMake).
- **`detect-toolchains`**: Automatically detect system toolchains and create definitions in `toolchains/`.
- **`add-config <config_name>`**: Add a build configuration.
- **`remove-config <config_name>`**: Remove a build configuration.

## YAML Formats

### `cbuild_workspace.yml`

Located at the root of the workspace.

```yaml
cmake_binary: "/usr/bin/cmake"    # Optional: Path to cmake binary
cxx_version: "20"                 # Default C++ standard for the workspace
configurations: ["Debug", "Release"] # Default build configurations

targets:
  <sourcename>:
    project_type: "cmake"         # Currently only "cmake" is supported
    depends: ["dep1", "dep2/sub"] # List of dependencies
    cmake_package_name: "Name"    # Optional: For CMake's find_package()
    cxx_standard: "17"            # Optional: Override workspace C++ version
    staged: true                  # Optional: Use staging for this target
    extra_cmake_configure_args: ["-DFOO=BAR"] # Optional: Extra args for CMake
```

### Toolchain `toolchain.yml`

Located in `toolchains/<toolchain_name>/toolchain.yml`.

```yaml
target_arch: "x64"
target_system: "linux"
cmake_toolchain:
  <host_key>:
    cmake_toolchain_file: "path/to/toolchain.cmake"
```

The `<host_key>` typically follows the format `host-<os>-<arch>` (e.g., `host-linux-x64`).



