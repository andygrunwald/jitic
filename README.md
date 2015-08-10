# jitic

*jitic* - the **JI**RA **Ti**cket **C**hecker - checks the existence of one or more ticket in the [JIRA](https://www.atlassian.com/software/jira) issue tracker. If the tickets exists we will shutdown with exit code 0. Otherwise with 1.

## Usage

TODO

## Use cases

### Subversion "pre-commit" hook

See [Implementing Repository Hooks](http://svnbook.red-bean.com/en/1.7/svn.reposadmin.create.html#svn.reposadmin.create.hooks) and [pre-commit](http://svnbook.red-bean.com/en/1.7/svn.ref.reposhooks.pre-commit.html).

How a pre-commit hook can look like:
```sh
#!/bin/sh

REPOS="$1"
TXN="$2"

# Get the commit message
SVNLOOK=/usr/bin/svnlook
COMMIT_MSG=$($SVNLOOK log -t "$TXN" "$REPOS")

JITIC=/usr/bin/jitic
JIRA_URL="https://jira.example.org/"
JIRA_USERNAME="JIRA-API"
JIRA_PASSWORD="SECRET-PASSWORD"

# Exit on all errors.
set -e

# Auth against JIRA and check if the ticket(s) exists
$JITIC -url="$JIRA_URL" -user="$JIRA_USERNAME" -pass="$JIRA_PASSWORD" -ticket-message="$COMMIT_MSG"

# All checks passed, so allow the commit.
exit 0
```

### Git "pre-receive" hook

See [Customizing Git - Git Hooks](https://git-scm.com/book/it/v2/Customizing-Git-Git-Hooks).

How a pre-receive hook can look like:
```sh
#!/bin/sh


```

## License

This project is released under the terms of the [MIT license](http://en.wikipedia.org/wiki/MIT_License).
