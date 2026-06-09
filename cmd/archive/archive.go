package archive

import (
	"log"
	"sort"

	"github.com/mrcyjanek/simplybs/crash"
	"github.com/mrcyjanek/simplybs/utils"
	"github.com/mrcyjanek/simplybs/utils/download"
)

func Archive() {
	sources, err := utils.LoadSourcesFile()
	crash.Handle(err)

	repoURLs := make([]string, 0, len(sources.Repositories))
	for url := range sources.Repositories {
		repoURLs = append(repoURLs, url)
	}
	sort.Strings(repoURLs)

	for i, url := range repoURLs {
		log.Printf("Ensuring git bundle %d/%d: %s", i+1, len(repoURLs), url)
		bundlePath := utils.SourcePathForGitURL(url)
		if err := utils.EnsureGitBundle(bundlePath, url); err != nil {
			log.Fatalf("Failed to ensure git bundle for %s: %v", url, err)
		}
	}

	downloads := utils.UniqueDownloads(sources.Downloads)
	for i, entry := range downloads {
		log.Printf("Ensuring download %d/%d: %s", i+1, len(downloads), entry.URL)
		path := utils.SourcePathForFileURL(entry.URL)
		if err := download.EnsureDownloadFile("archive", path, entry.URL, entry.Sha256); err != nil {
			log.Fatalf("Failed to download %s: %v", entry.URL, err)
		}
	}

	log.Printf("Archive complete: %d git repos, %d unique downloads", len(repoURLs), len(downloads))
}
