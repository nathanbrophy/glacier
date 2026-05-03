---
slug: exit:130
title: "Exit code 130: interrupted"
category: exit-code
see_also: ["exit:143"]
---
The process was interrupted by SIGINT (Ctrl-C). glacier handles SIGINT gracefully via context cancellation; any in-progress writes are flushed before exit.
