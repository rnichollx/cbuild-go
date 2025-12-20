# cbuild  csetup

cbuild is a tool for building distributions. You can
think of it as a way to build many projects
simultaneously. cbuild works locally to build a cbuild
workspace.

csetup is a related tool to download or otherwise set up
cbuild workspaces. For example, you might create a cbuild
workspace like so:

```
csetup create ./my_workspace --from-url http://example.com/my_project/git/latest/csetuplists.yml 
```
Or using a csetup file that you provide:

```
csetup create ./my_workspace /srv/dev_shared/foo_project_setup.json
```

cbuild workspaces have the following structure:

```
workspace/
  
  cbuid_workspace.yml
  toolchains/<toolchain_name>/
  sources/<sourcename>/
  buildspaces/<sourcename>/<toolchain>/<config>
  exports/<sourcename>/<toolchain>/<config>
```

## CSetup

### `csetup git-clone <repo_url> <dest_name>` 

## cbuild conepts

### sources directory

The directory in which to use as the base of relative source paths.

### buildspace directory

The directory to store object files, build system files, and other temporary files for building, including any files preserved between builds for incremental builds.

### toolchains directory

The directory to look for build toolchain definition files.

### exports directory

The directory to store any files that result from the build process.

## cbuild settings

### cbuild_workspace.yml

This file defines the settings for the cbuild workspace.

Fields:

#### `targets`

The targets field defines a list of targets to build. Each target is defined in its own directory under the sources
directory.

A target object has the following fields:

- `depends`: An array of list of named dependencies for the project.
- `project_type`: The type of project. Example values are "CMake".
- `cmake_package_name`: The name of the package if it's a CMake Package.

A dependency has the form of `<sourcename>[/<subtarget>]`, where `<sourcename>` is the name of a source target,
and optionally `<subtarget>` is a subtarget of that build system.



