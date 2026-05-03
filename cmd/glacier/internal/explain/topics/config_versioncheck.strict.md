---
slug: config:versioncheck.strict
title: "Config key: versioncheck.strict"
category: config-key
see_also: ["config:versioncheck.enabled", "exit:68"]
---
When true, glacier version exits with code 68 if the GitHub Releases API is unreachable. Without this, unreachability exits 0 with an (offline) annotation.

Default: false
Env override: GLACIER__VERSIONCHECK__STRICT
