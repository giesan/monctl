# Community issue and pull request templates

## Scope

Add GitHub-recognized issue templates for bug reports and feature requests and
one pull request template. Do not add a code of conduct or a private conduct
reporting channel. Preserve the existing contribution and security policies.

## Templates

Use short Markdown templates under `.github/ISSUE_TEMPLATE/` and
`.github/PULL_REQUEST_TEMPLATE.md`. The bug report asks for the Linux
distribution, `monctl --version`, steps to reproduce, expected and actual
results, and relevant diagnostics with secrets removed. The feature request
asks what problem it solves and what behavior is desired. Both issue templates
point suspected vulnerabilities to the existing private reporting route in
`SECURITY.md` rather than public issues. The PR template asks for a summary,
related issue (if any), tests run and limitations, and a screenshot when the
desktop window changes, consistent with `CONTRIBUTING.md`. Avoid making issue
fields mandatory or blocking contributors who lack hardware for a test.

## Verification

Check that GitHub's conventional paths and Markdown front matter are valid,
links to security and contribution guidance resolve, and `git diff --check`
passes. These documentation-only changes do not need application tests. Do not
commit the unrelated untracked `GITHUB-PUBLICATION-LOCAL.md` file or publish a
release as part of this change.
