package utils

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"sort"
	"strings"
)

//go:embed sources_ignored_refs.txt
var sourcesIgnoredRefsFile string

var sourcesIgnoredRefs map[string]struct{}

func init() {
	sourcesIgnoredRefs = make(map[string]struct{})
	for _, line := range strings.Split(sourcesIgnoredRefsFile, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		sourcesIgnoredRefs[line] = struct{}{}
	}
}

func IsIgnoredGitRef(ref string) bool {
	_, ignored := sourcesIgnoredRefs[ref]
	return ignored
}

func FilterIgnoredGitRefs(s *SourcesFile) {
	for url, repoInfo := range s.Repositories {
		filtered := make([]string, 0, len(repoInfo.Refs))
		for _, ref := range repoInfo.Refs {
			if !IsIgnoredGitRef(ref) {
				filtered = append(filtered, ref)
			}
		}
		if len(filtered) == 0 {
			delete(s.Repositories, url)
			continue
		}
		repoInfo.Refs = filtered
		s.Repositories[url] = repoInfo
	}
}

func NewSourcesFile() *SourcesFile {
	return &SourcesFile{
		Repositories: make(map[string]RepositoryInfo),
	}
}

func AddGitRef(s *SourcesFile, url, ref string) bool {
	if url == "" || ref == "" || IsIgnoredGitRef(ref) {
		return false
	}
	repoInfo, exists := s.Repositories[url]
	if exists {
		for _, existing := range repoInfo.Refs {
			if existing == ref {
				return false
			}
		}
		repoInfo.Refs = append(repoInfo.Refs, ref)
		s.Repositories[url] = repoInfo
		return true
	}
	s.Repositories[url] = RepositoryInfo{
		Refs: []string{ref},
	}
	return true
}

func downloadEntryKey(entry DownloadEntry) string {
	data, err := json.Marshal(entry)
	if err != nil {
		return entry.Kind + "\x00" + entry.URL + "\x00" + entry.Sha256 + "\x00" + entry.Path
	}
	return string(data)
}

func AddDownload(s *SourcesFile, entry DownloadEntry) bool {
	if entry.Kind == "" || entry.URL == "" || entry.Sha256 == "" {
		return false
	}
	key := downloadEntryKey(entry)
	for _, existing := range s.Downloads {
		if downloadEntryKey(existing) == key {
			return false
		}
	}
	s.Downloads = append(s.Downloads, entry)
	return true
}

func downloadMapFields(download map[string]interface{}) (kind, url, sha256 string, ok bool) {
	kind, _ = download["kind"].(string)
	if kind == "" || kind == "none" {
		return "", "", "", false
	}
	url, _ = download["url"].(string)
	sha256, _ = download["sha256"].(string)
	return kind, url, sha256, true
}

func MergeDownloadMap(s *SourcesFile, download map[string]interface{}) {
	kind, url, sha256, ok := downloadMapFields(download)
	if !ok {
		return
	}

	if kind == "git" {
		AddGitRef(s, url, sha256)
		return
	}

	if url == "" || sha256 == "" {
		return
	}

	entry := DownloadEntry{
		Kind:   kind,
		URL:    url,
		Sha256: sha256,
	}
	if path, ok := download["path"].(string); ok && path != "" {
		entry.Path = path
	}
	AddDownload(s, entry)
}

func MergeDownloadsFromJSON(s *SourcesFile, content []byte) {
	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return
	}

	downloads, ok := data["download"].([]interface{})
	if !ok {
		return
	}

	for _, downloadInterface := range downloads {
		download, ok := downloadInterface.(map[string]interface{})
		if !ok {
			continue
		}
		MergeDownloadMap(s, download)
	}
}

func SortSourcesFile(s *SourcesFile) {
	for url, repoInfo := range s.Repositories {
		sort.Strings(repoInfo.Refs)
		s.Repositories[url] = repoInfo
	}

	sort.Slice(s.Downloads, func(i, j int) bool {
		a, b := s.Downloads[i], s.Downloads[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.URL != b.URL {
			return a.URL < b.URL
		}
		if a.Sha256 != b.Sha256 {
			return a.Sha256 < b.Sha256
		}
		return a.Path < b.Path
	})
}

func MarshalSourcesFile(s *SourcesFile) ([]byte, error) {
	FilterIgnoredGitRefs(s)
	SortSourcesFile(s)

	var buf bytes.Buffer
	if err := NewIndentedEncoder(&buf).Encode(s); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}

func CountGitRefs(s *SourcesFile) int {
	total := 0
	for _, repoInfo := range s.Repositories {
		total += len(repoInfo.Refs)
	}
	return total
}

func SourcesFileFromRepositoriesOnlyJSON(content []byte) (*SourcesFile, error) {
	var sources SourcesFile
	if err := json.Unmarshal(content, &sources); err != nil {
		return nil, err
	}
	if sources.Repositories == nil {
		sources.Repositories = make(map[string]RepositoryInfo)
	}
	return &sources, nil
}

func AllGitRefPairs(s *SourcesFile) [][2]string {
	pairs := make([][2]string, 0)
	for url, repoInfo := range s.Repositories {
		for _, ref := range repoInfo.Refs {
			pairs = append(pairs, [2]string{url, ref})
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i][0] != pairs[j][0] {
			return pairs[i][0] < pairs[j][0]
		}
		return pairs[i][1] < pairs[j][1]
	})
	return pairs
}

func ContainsGitRef(s *SourcesFile, url, ref string) bool {
	repoInfo, exists := s.Repositories[url]
	if !exists {
		return false
	}
	for _, existing := range repoInfo.Refs {
		if existing == ref {
			return true
		}
	}
	return false
}

func UniqueDownloads(entries []DownloadEntry) []DownloadEntry {
	seen := make(map[string]bool)
	unique := make([]DownloadEntry, 0, len(entries))

	for _, entry := range entries {
		key := entry.URL + "\x00" + entry.Sha256
		if seen[key] {
			continue
		}
		seen[key] = true
		unique = append(unique, entry)
	}

	sort.Slice(unique, func(i, j int) bool {
		a, b := unique[i], unique[j]
		if a.URL != b.URL {
			return a.URL < b.URL
		}
		return a.Sha256 < b.Sha256
	})

	return unique
}

func DownloadEntryFromMap(download map[string]interface{}) (DownloadEntry, bool) {
	kind, url, sha256, ok := downloadMapFields(download)
	if !ok || kind == "git" {
		return DownloadEntry{}, false
	}
	if url == "" || sha256 == "" {
		return DownloadEntry{}, false
	}
	entry := DownloadEntry{
		Kind:   kind,
		URL:    url,
		Sha256: sha256,
	}
	if path, ok := download["path"].(string); ok && strings.TrimSpace(path) != "" {
		entry.Path = path
	}
	return entry, true
}
