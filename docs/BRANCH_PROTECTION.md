# Branch Protection

## Recommended Settings

Use the following settings to protect the default branch (for example, `master`):

- Require a pull request before merging
- Require at least one approval
- Require status checks to pass before merging
- Dismiss stale approvals when new commits are pushed
- Prevent force pushes
- Prevent branch deletion
- Optionally require review from code owners

## How to Enable in GitHub

1. Open the repository on GitHub.
2. Go to `Settings`.
3. Select `Branches` from the left navigation.
4. Under `Branch protection rules`, choose `Add rule`.
5. In `Branch name pattern`, enter the default branch name (for example, `master`).
6. Enable the settings listed above.
7. Click `Create` or `Save changes`.
