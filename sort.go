package main

import "github.com/Masterminds/semver/v3"

type ImageTags []string

func (t ImageTags) Len() int {
	return len(t)
}

func (t ImageTags) Less(i, j int) bool {
	vi, erri := semver.NewVersion(t[i])
	vj, errj := semver.NewVersion(t[j])
	switch {
	case erri == nil && errj == nil:
		return vi.LessThan(vj)
	case erri == nil && errj != nil:
		return true
	case erri != nil && errj == nil:
		return false
	default:
		return t[i] < t[j]
	}
}

func (t ImageTags) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
