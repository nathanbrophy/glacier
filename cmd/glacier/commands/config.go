// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"time"

	"github.com/nathanbrophy/glacier/conf"
)

// sdkConfig is the SDK configuration type registered via conf.
type sdkConfig struct {
	GitHub       gitHubConfig       `json:"github"`
	VersionCheck versionCheckConfig `json:"versioncheck"`
	Banner       bannerConfig       `json:"banner"`
	Palette      map[string]string  `json:"palette,omitempty"`
	Telemetry    bool               `json:"-"` // always false; ignored
}

type gitHubConfig struct {
	Repo string `json:"repo"` // default "nathanbrophy/glacier"
}

type versionCheckConfig struct {
	CacheTTL time.Duration `json:"cache_ttl"` // default 24h
	Enabled  bool          `json:"enabled"`   // default true
	Strict   bool          `json:"strict"`    // default false
}

type bannerConfig struct {
	ShowOnHelp bool `json:"show_on_help"` // default true
}

// sdkCfg is the package-level accessor for SDK configuration.
// conf.Register panics on duplicate paths, so this init is side-effect-safe
// as long as nothing else registers "glacier" in the same process.
var sdkCfg = conf.Register("glacier", sdkConfig{
	GitHub:       gitHubConfig{Repo: "nathanbrophy/glacier"},
	VersionCheck: versionCheckConfig{CacheTTL: 24 * time.Hour, Enabled: true},
	Banner:       bannerConfig{ShowOnHelp: true},
})
