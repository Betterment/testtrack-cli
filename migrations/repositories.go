package migrations

import "sort"

// Repository is a map of migrations indexed by migration version
type Repository map[string]IMigration

// SortedVersions sorts and returns the migration versions in a repo because
// maps don't preserve order in go
func (m *Repository) SortedVersions() []string {
	versions := make([]string, 0, len(*m))

	for version := range *m {
		versions = append(versions, version)
	}

	sort.Strings(versions)

	return versions
}
