# gh-org-admin-promote

GitHub CLI extension to promote an enterprise admin to an organization admin for all orgs in the enterprise.  This is an API-first replacement of `ghe-org-admin-promote` on GitHub Enterprise Server.  It also outputs an inventory of all organizations in the enterprise as a CSV file.

Should work on [all supported versions](https://docs.github.com/en/enterprise-server@latest/admin/all-releases#releases-of-github-enterprise-server) of GitHub Enterprise Server, as well as GitHub Enterprise Cloud.

## Permissions check

> [!IMPORTANT]
> This requires the `admin:enterprise` and `admin:org` scopes, which are only available to enterprise owners and not default for logging in to the gh cli.

Run `ghe auth status` to check your permissions.  You should see `admin:enterprise` and `admin:org` in the list of scopes.

```console
$ gh auth status

ghes-test-instance.com
  ✓ Logged in to ghes-test-instance.com account some-natalie (keyring)
  - Active account: true
  - Git operations protocol: https
  - Token: gho_************************************
  - Token scopes: 'admin:enterprise', 'admin:org', 'gist', 'repo', 'workflow'
```

If you don't, do the following to add the right scopes:

```console
gh auth refresh -s admin:enterprise -s admin:org -h ghes-test-instance.com
```

## Installation

```console
gh extension install some-natalie/gh-org-admin-promote
```

## Usage

```console
$ export GH_HOST=ghes-test-instance.com  # option for GHES, defaults to github.com

$ gh org-admin-promote enterprise-name

Getting total count of organizations in github...
Total count of organizations in github: 4
Getting list of organizations in github...
Promoting user to admin for testorg-00002...
User promoted to admin for testorg-00002
Promoting user to admin for testorg-00003...
User promoted to admin for testorg-00003
```

## Limitations

This will promote you to own all organizations, but it will not capture anything in a user-namespaced repository (e.g. `some-natalie/gh-org-admin-promote`).  If you need reporting on all of these, for GHES, use the [all_repositories.csv report](https://docs.github.com/en/enterprise-server@latest/admin/administering-your-instance/administering-your-instance-from-the-web-ui/site-admin-dashboard#reports) to get a list.
