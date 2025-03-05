package util

import (
	"github.com/Masterminds/semver/v3"
)

// FindHighestVersion finds the highest semver version in a list of versions
func FindHighestVersion(versions []string) string {
	var highestVersion *semver.Version

	for _, str := range versions {
		v := semver.MustParse(str)

		if highestVersion == nil || v.GreaterThan(highestVersion) {
			highestVersion = v
		}
	}

	return highestVersion.String()
}
