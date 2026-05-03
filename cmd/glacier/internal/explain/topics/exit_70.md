---
slug: exit:70
title: "Exit code 70: subprocess failure"
category: exit-code
see_also: ["exit:66", "exit:65"]
---
A subprocess launched by glacier (go test, staticcheck) exited non-zero for a reason unrelated to test or lint findings. Check stderr for the subprocess output.
