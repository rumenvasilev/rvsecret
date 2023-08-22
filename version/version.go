package version

import (
	"fmt"
)

const (
	Name = "flower"
)

// AppVersionMajor is the major revision number
const AppVersionMajor = "0"

// AppVersionMinor is the minor revision number
const AppVersionMinor = "1"

// AppVersionPatch is the patch version
const AppVersionPatch = "0"

// AppVersionPre ...
const AppVersionPre = ""

// AppVersionBuild should be empty string when releasing
const AppVersionBuild = ""

// AppVersion generates a usable version string
func AppVersion() string {
	return fmt.Sprintf("%s.%s.%s%s%s", AppVersionMajor, AppVersionMinor, AppVersionPatch, AppVersionPre, AppVersionBuild)
}

// UserAgent set the browser user agent when required.
var UserAgent = fmt.Sprintf("%s v%s", Name, AppVersion())
