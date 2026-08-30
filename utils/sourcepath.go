package utils

import (
	"path/filepath"
	"strings"

	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/utils/download"
)

func SourcePathForGitURL(url string) string {
	urlPath, err := download.URLToPath(url)
	if err != nil {
		return filepath.Join(host.DataDirRoot(), "source", filepath.Base(url)+".bundle")
	}
	urlPath = strings.TrimSuffix(urlPath, ".git")
	return filepath.Join(host.DataDirRoot(), "source", urlPath+".bundle")
}

func SourcePathForFileURL(url string) string {
	urlPath, err := download.URLToPath(url)
	if err != nil {
		return filepath.Join(host.DataDirRoot(), "source", filepath.Base(url))
	}
	return filepath.Join(host.DataDirRoot(), "source", urlPath)
}
