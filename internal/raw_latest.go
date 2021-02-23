// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

import (
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
)

// RawLatestInfo describes the "raw" latest version of a module:
// the latest version without considering retractions or the like.
// The go.mod file of the raw latest version establishes whether
// the module is deprecated, and what versions are retracted.
type RawLatestInfo struct {
	ModulePath string
	Version    string
	GoModFile  *modfile.File
}

// PopulateModule uses the RawLatestInfo to populate fields of the given module.
func (r *RawLatestInfo) PopulateModule(m *Module) {
	m.Deprecated, m.DeprecationComment = isDeprecated(r.GoModFile)
	m.Retracted, m.RetractionRationale = isRetracted(r.GoModFile, m.Version)
}

// isDeprecated reports whether the go.mod deprecates this module.
// It looks for "Deprecated" comments in the line comments before and next to
// the module declaration. If it finds one, it returns true along with the text
// after "Deprecated:". Otherwise it returns false, "".
func isDeprecated(mf *modfile.File) (bool, string) {
	const prefix = "Deprecated:"

	if mf.Module == nil {
		return false, ""
	}
	for _, comment := range append(mf.Module.Syntax.Before, mf.Module.Syntax.Suffix...) {
		text := strings.TrimSpace(strings.TrimPrefix(comment.Token, "//"))
		if strings.HasPrefix(text, prefix) {
			return true, strings.TrimSpace(text[len(prefix):])
		}
	}
	return false, ""
}

// isRetracted reports whether the go.mod file retracts the version.
// If so, it returns true along with the rationale for the retraction.
func isRetracted(mf *modfile.File, resolvedVersion string) (bool, string) {
	for _, r := range mf.Retract {
		if semver.Compare(resolvedVersion, r.Low) >= 0 && semver.Compare(resolvedVersion, r.High) <= 0 {
			return true, r.Rationale
		}
	}
	return false, ""
}