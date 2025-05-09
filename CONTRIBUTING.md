## 🛠️ Hotfix vs Future Sprint Decision Table

| Situation                                                                 | Future Sprint | Hotfix Cycle | Notes                                                             |
|---------------------------------------------------------------------------|---------------|--------------|-------------------------------------------------------------------|
| 🧱 Critical system crash (agent or dashboard)                              | ❌            | ✅           | Blocking functionality or causes data loss. Fix ASAP.            |
| ❌ Login/authentication failure                                           | ❌            | ✅           | Prevents access to the app — always hotfix.                      |
| 🐞 Agent fails to install or register                                     | ❌            | ✅           | Blocks core functionality — fix immediately.                     |
| 🔐 Security vulnerability or exploit risk                                | ❌            | ✅           | Especially if public-facing or high severity.                    |
| 🧪 Data corruption bug (e.g. overwriting wrong config)                   | ❌            | ✅           | High risk to users/data.                                         |
| 🕐 Dashboard performance issues (slow load, lag)                          | ✅            | ❌           | Non-blocking — prioritize in future sprint.                      |
| 🧹 UI polish or layout issues (styling, alignment)                        | ✅            | ❌           | Cosmetic only.                                                   |
| 🛠️ Feature enhancement or new module                                     | ✅            | ❌           | Always planned as part of a future sprint.                       |
| 🧪 Test failures (unit/integration) in dev/staging                        | ✅            | ❌           | Not user-facing — fix in next sprint.                            |
| 📦 Deployment script issue in staging or CI/CD                           | ✅ or ✅\*     | ✅           | Fix in sprint unless blocking production deploys.                |
| 🧑‍💻 Code refactoring or tech debt cleanup                                | ✅            | ❌           | Valuable but non-urgent — schedule later.                        |
| 🧩 Plugin/module fails in edge case                                      | ✅            | ❌ (unless widespread) | Defer unless severely affecting users.                   |
| 🔁 Script execution logs not storing correctly                            | ✅            | ❌ (unless data critical) | Fix in sprint unless data loss is critical.              |
| 📤 Remote control intermittently fails                          | ✅            | ✅ (if mission-critical) | Depends on how essential the feature is.                 |
| 📉 Metrics not updating on dashboard (non-critical)                       | ✅            | ❌           | Fix in a future sprint.                                          |
| 🧱 Install script fails on supported OS (e.g., Ubuntu)                    | ❌            | ✅           | Breaks usability — hotfix.                                       |
| 🛡️ Dependency with published CVE/security advisory                       | ✅ or ✅\*     | ✅ (if actively exploited) | Evaluate severity — patch if necessary.                |

\* Depends on context and severity.
