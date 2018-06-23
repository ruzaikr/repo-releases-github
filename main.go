package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/google/go-github/github"
)

const RELEASES_PER_PAGE = 100

//--- custom sorting ---

type ByMajorMinorPatch []*semver.Version

func (v ByMajorMinorPatch) Len() int {
	return len(v)
}

func (v ByMajorMinorPatch) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v ByMajorMinorPatch) Less(i, j int) bool {
	a := v[i]
	b := v[j]

	if a == nil || b == nil {
		return false
	}

	if a.Major != b.Major {
		return a.Major > b.Major
	}

	if a.Minor != b.Minor {
		return a.Minor > b.Minor
	}

	return a.Patch > b.Patch
}

//--- end of custom sorting ---

// greaterThanMin returns whether a version (first param) is greater than the min version (second param)
func greaterThanMin(v *semver.Version, minVersion *semver.Version) bool {
	if v.Major > minVersion.Major {
		return true
	} else if v.Major == minVersion.Major {
		if v.Minor > minVersion.Minor {
			return true
		} else if v.Minor == minVersion.Minor {
			if v.Patch >= minVersion.Patch {
				return true
			}
		}
	}
	return false
}

// LatestVersions returns a sorted slice with the highest version as its first element and the highest version of the smaller minor versions in a descending order
func LatestVersions(releases []*semver.Version, minVersion *semver.Version) []*semver.Version {
	sort.Sort(ByMajorMinorPatch(releases))
	versionSlice := make([]*semver.Version, 0)
	var lastVersion *semver.Version

	for _, v := range releases {
		if v == nil {
			continue
		}

		if v.PreRelease != "" { // Skip pre-releases
			continue
		}

		if greaterThanMin(v, minVersion) {
			if len(versionSlice) > 0 {
				if v.Major == lastVersion.Major && v.Minor == lastVersion.Minor {
					// This means that 'v' is a smaller minor version than the last one that was appended
					continue
				}
			}
			versionSlice = append(versionSlice, v)
			lastVersion = v
		} else {
			break
		}
	}

	return versionSlice
}

type Input struct {
	Owner      string
	Repo       string
	MinVersion *semver.Version
}

// readInputFromFile opens a file 'input' in the project root and returns a built []Input
func readInputFromFile(path string) []Input {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal("Failed to open text file named 'input' in project root. Details: " + err.Error())
	}
	defer file.Close()

	repos := make([]Input, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "repository,min_version" { // Skip the first line of the input file
			continue
		}

		// Create a new Input struct with the values from the read line of string
		i := Input{}
		s1 := strings.Split(line, "/")
		i.Owner = s1[0]
		s2 := strings.Split(s1[1], ",")
		i.Repo = s2[0]
		i.MinVersion = semver.New(s2[1])

		repos = append(repos, i)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return repos
}

// getReleasesForRepoFromGithub returns a full list (all pages) of releases for a particular owner/repo
func getReleasesForRepoFromGithub(client *github.Client, repoInput *Input) ([]*github.RepositoryRelease, *github.Rate,
	error) {
	ctx := context.Background()
	opt := &github.ListOptions{Page: 1, PerPage: RELEASES_PER_PAGE}

	loop := true
	releases := make([]*github.RepositoryRelease, 0)

	for loop {
		releasesPerPage, resp, err := client.Repositories.ListReleases(ctx, repoInput.Owner, repoInput.Repo, opt)
		if err != nil {
			return releases, &resp.Rate, err
		}

		releases = append(releases, releasesPerPage...)
		loop = resp.NextPage > 0
		opt.Page++
	}

	return releases, nil, nil
}

// Validate version string using regexp
func validVersionString(versionString string) bool {
	reVersion := regexp.MustCompile(`^\d{1,}.\d{1,}.\d{1,}$`)
	reVersionWithPreRelease := regexp.MustCompile(`^\d{1,}.\d{1,}.\d{1,}-[0-9A-Za-z-]*(.[0-9A-Za-z-]*)*$`)

	return reVersion.MatchString(versionString) || reVersionWithPreRelease.MatchString(versionString)
}

// Here we implement the basics of communicating with github through the library as well as printing the version
// You will need to implement LatestVersions function as well as make this application support the file format outlined in the README
// Please use the format defined by the fmt.Printf line at the bottom, as we will define a passing coding challenge as one that outputs
// the correct information, including this line
func main() {
	path := os.Args[1]
	repos := readInputFromFile(path)

	client := github.NewClient(nil)

	for _, repoInput := range repos {
		releases, rate, err := getReleasesForRepoFromGithub(client, &repoInput)
		if err != nil {
			if rate != nil && rate.Remaining == 0 {
				log.Fatalf("Reached Github rate limit for unauthorized requests. Details: %v.", err)
			}

			log.Printf("Failed to retrieve all releases for %s/%s. Details: %v.", repoInput.Owner,
				repoInput.Repo, err)
			// TODO: Is it better to stop here?
			continue

		}

		allReleases := make([]*semver.Version, len(releases))
		for i, release := range releases {
			versionString := *release.TagName
			if versionString[0] == 'v' {
				versionString = versionString[1:]
			}

			if validVersionString(versionString) {
				allReleases[i] = semver.New(versionString)
			} // invalid version strings will be ignored else they will cause semver to panic

		}

		versionSlice := LatestVersions(allReleases, repoInput.MinVersion)
		fmt.Printf("latest versions of %s/%s: %s\n", repoInput.Owner, repoInput.Repo, versionSlice)
	}

}
