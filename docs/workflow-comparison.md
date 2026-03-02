# Workflow comparison: Terraform plan/apply with GitHub Actions and Atlantis

This page compares three ways to run Terraform or OpenTofu plan and apply in a pull-request workflow. The main idea: **apply-before-merge** keeps your default branch (`main`) fully executable, because apply runs on the PR and you only merge when it succeeds. In a "normal" workflow where apply runs after merge, code on `main` is not guaranteed to be executable—errors surface after merge and require follow-up PRs.

---

## Normal Terraform + GitHub Actions workflow

Apply runs *after* merge (e.g. on push to `main` or a post-merge job). Code on `main` is **not** guaranteed to be fully executable.

1. Dev changes `.tf` → opens PR
2. Plan executed (on PR)
3. PR approved → merge
4. Apply executed (on `main`)
5. **Error** → open another PR to fix
6. Plan executed → PR approved → merge → Apply executed
7. **Error** again → repeat…

Each failure lands broken code on `main` and forces a new PR cycle.

---

## Neptune workflow

Apply runs on the PR when you trigger it (e.g. comment `@neptbot apply`). You merge only after apply succeeds. All code on `main` stays **fully executable**. Everything runs **inside GitHub Actions**—no separate servers.

1. Dev changes `.tf` → opens PR
2. Plan executed (GitHub Actions)
3. PR approved → comment `@neptbot apply`
4. Apply executed (GitHub Actions)
5. **Error** → push changes to the same PR
6. `@neptbot apply` again → Apply executed
7. Merge when apply succeeds (all within GitHub Actions)

---

## Atlantis workflow

Same apply-before-merge idea as Neptune: plan on PR, apply on PR (e.g. comment `atlantis apply`), merge when apply succeeds. **Difference**: execution runs on a **separate (self-hosted) server**, not in GitHub Actions.

1. Dev changes `.tf` → opens PR
2. Plan executed (Atlantis server)
3. PR approved → comment `atlantis apply`
4. Apply executed (Atlantis server)
5. **Error** → push changes to the same PR
6. `atlantis apply` again → Apply executed
7. Merge when apply succeeds (execution stays on the Atlantis server)

---

## Comparison at a glance

| | Normal Terraform + GHA | Neptune | Atlantis |
| --- | --------------------- | ------- | -------- |
| **Apply before merge?** | No (apply after merge) | Yes | Yes |
| **Main branch executable?** | No | Yes | Yes |
| **Where plan/apply run?** | GitHub Actions | GitHub Actions | Self-hosted server |
| **Extra infrastructure?** | No | No | Yes (Atlantis server) |

**When to choose Neptune**: You want apply-before-merge and to keep everything in GitHub Actions—no extra servers or self-hosted runners to operate.

**When Atlantis may fit**: You already run Atlantis or prefer a dedicated self-hosted service for Terraform runs; the workflow is similar, but execution is outside GitHub.
