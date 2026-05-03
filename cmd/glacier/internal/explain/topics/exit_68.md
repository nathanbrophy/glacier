---
slug: exit:68
title: "Exit code 68: version check unreachable"
category: exit-code
see_also: ["exit:67", "config:versioncheck.strict"]
---
glacier version --check could not reach the GitHub Releases API and --strict was set. Without --strict the command exits 0 with an (offline) annotation.

Run glacier explain config:versioncheck.strict for the config key.
