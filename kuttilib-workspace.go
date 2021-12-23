package kuttilib

import "github.com/kuttiproject/workspace"

// SetWorkspace sets the kutti workspace to the path specified.
// Config and Cache directories are set as subdirectories under the specified path,
// called kutti-config and kutti-cache respectively.
func SetWorkspace(workspacepath string) error {
	err := workspace.Set(workspacepath)
	if err == nil {
		setworkspaceconfigmanager()
	}
	return err
}

// ResetWorkspace resets the kutti workspace to the default location.
// Config and Cache directories are set as subdirectories called kutti
// under the current user's config and cache locations respectively.
func ResetWorkspace() {
	workspace.Reset()
	setworkspaceconfigmanager()
}
