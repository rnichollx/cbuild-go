package ccommon

import (
	"context"
	"errors"
)

type CodeSource struct {
	Git   *GitSource `yaml:"git,omitempty"`
	Local string     `yaml:"local,omitempty"`
}

func (cs *CodeSource) hasGit() bool {
	return cs.Git != nil
}

func (cs *CodeSource) hasLocal() bool {
	return cs.Local != ""
}

func (cs *CodeSource) Validate() error {
	if !cs.hasGit() && !cs.hasLocal() {

		return errors.New("code source must have either git or local defined")
	}
	if cs.hasGit() && cs.hasLocal() {
		return errors.New("code source cannot have both git and local defined")
	}
	return nil
}

func (cs *CodeSource) ValidateWeb() error {
	if err := cs.Validate(); err != nil {
		return err
	}
	if cs.hasLocal() {
		return errors.New("local code sources are not allowed in remote context")
	}
	return nil
}

func (cs *CodeSource) From() string {

	if cs.Git != nil {
		if cs.Git.Revision != nil {
			return cs.Git.Repository + "@" + *cs.Git.Revision
		}
		return cs.Git.Repository
	}
	return cs.Local
}

func (ws *WorkspaceContext) Get(ctx context.Context, name string, source CodeSource) error {
	if source.hasGit() {
		return ws.GetFromGit(ctx, name, *source.Git)
	}
	return errors.New("unsupported source type")
}
