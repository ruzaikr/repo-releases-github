package main

import (
	"testing"

	"github.com/coreos/go-semver/semver"
	"github.com/stretchr/testify/assert"
)

func stringToVersionSlice(stringSlice []string) []*semver.Version {
	versionSlice := make([]*semver.Version, len(stringSlice))
	for i, versionString := range stringSlice {
		versionSlice[i] = semver.New(versionString)
	}
	return versionSlice
}

func versionToStringSlice(versionSlice []*semver.Version) []string {
	stringSlice := make([]string, len(versionSlice))
	for i, version := range versionSlice {
		stringSlice[i] = version.String()
	}
	return stringSlice
}

func TestLatestVersions(t *testing.T) {
	testCases := []struct {
		versionSlice   []string
		expectedResult []string
		minVersion     *semver.Version
	}{
		{
			versionSlice:   []string{"1.8.11", "1.9.6", "1.10.1", "1.9.5", "1.8.10", "1.10.0", "1.7.14", "1.8.9", "1.9.5"},
			expectedResult: []string{"1.10.1", "1.9.6", "1.8.11"},
			minVersion:     semver.New("1.8.0"),
		},
		{
			versionSlice:   []string{"1.8.11", "1.9.6", "1.10.1", "1.9.5", "1.8.10", "1.10.0", "1.7.14", "1.8.9", "1.9.5"},
			expectedResult: []string{"1.10.1", "1.9.6"},
			minVersion:     semver.New("1.8.12"),
		},
		{
			versionSlice:   []string{"1.10.1", "1.9.5", "1.8.10", "1.10.0", "1.7.14", "1.8.9", "1.9.5"},
			expectedResult: []string{"1.10.1"},
			minVersion:     semver.New("1.10.0"),
		},
		{
			versionSlice:   []string{"2.2.1", "2.2.0"},
			expectedResult: []string{"2.2.1"},
			minVersion:     semver.New("2.2.1"),
		},
		// Handle abnormal case where there are no releases after the min version
		{
			versionSlice:   []string{"2.2.6", "2.2.2", "2.4.8"},
			expectedResult: []string{},
			minVersion:     semver.New("2.6.1"),
		},
		// Handle abnormal case where the input does not contain any versions
		{
			versionSlice:   []string{},
			expectedResult: []string{},
			minVersion:     semver.New("2.6.1"),
		},
		// Implement more relevant test cases here, if you can think of any
	}

	test := func(versionData []string, expectedResult []string, minVersion *semver.Version) {
		stringSlice := versionToStringSlice(LatestVersions(stringToVersionSlice(versionData), minVersion))
		for i, versionString := range stringSlice {
			if versionString != expectedResult[i] {
				t.Errorf("Received %s, expected %s", stringSlice, expectedResult)
				return
			}
		}
	}

	for _, testValues := range testCases {
		test(testValues.versionSlice, testValues.expectedResult, testValues.minVersion)
	}
}

type VersionStringWithValidity struct {
	VersionString string
	ExpectedValidity bool
}

func TestValidVersionString(t *testing.T)  {
	vStrings := []VersionStringWithValidity{
		{"1.2.1", true},
		{"12.4.2", true},
		{"6.33.2", true},
		{"1.2.1-alpha.1", true},
		{"1.2.3-alpha.10.beta.0+build.unicorn.rainbow", true},
		{"0.2.1", true},
		{"hello", false},
		{"0.23.12.3", false},
		{"", false},
		{"v1.3.1", false},
	}

	for _, vs := range vStrings {
		assert.Equal(t, vs.ExpectedValidity, validVersionString(vs.VersionString),
			"Tested version string: %s", vs.VersionString)
	}
}