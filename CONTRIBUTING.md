# Contributing

Contributions are welcomed!

When contributing to this repository, please first discuss the change you wish to make via GitHub
issue before making a change.  This saves everyone from wasted effort in the event that the proposed
changes need some adjustment before they are ready for submission.

## Requirements

* go v1.17+
* goreleaser
* gofmt

## Testing

Run `make test` from the vergo directory.

## Changelog

Please ensure you update the `CHANGELOG.md`.

## Releasing

Only administrators of the Vergo repository can make new releases.

In order to release you will need to ensure you are on a commit prefixed with `[vergo:(major|minor|patch)-release]`.

Releases are automated with [goreleaser](https://goreleaser.com). To perform a release, and administrator must:

- Provide a Github access token with `repo` permissions as the GITHUB_TOKEN environment variable.
- Run `make release`.

## Pull Request Process

1. If your changes include multiple commits, please squash them into a single commit.  Stack Overflow
   and various blogs can help with this process if you're not already familiar with it.
2. Make sure to commit changes to vendor, ideally as a separate commit to any other code change.
3. Update the README.md where relevant.
4. You may merge the Pull Request in once you have the sign-off, or if you do not have permission to
   do that, you may request the reviewer to merge it for you.

## Contributor Code of Conduct

As contributors and maintainers of this project, and in the interest of fostering an open and
welcoming community, we pledge to respect all people who contribute through reporting issues,
posting feature requests, updating documentation, submitting pull requests or patches, and other
activities.

We are committed to making participation in this project a harassment-free experience for everyone,
regardless of level of experience, gender, gender identity and expression, sexual orientation,
disability, personal appearance, body size, race, ethnicity, age, religion, or nationality.

Examples of unacceptable behavior by participants include:

* The use of sexualized language or imagery
* Personal attacks
* Trolling or insulting/derogatory comments
* Public or private harassment
* Publishing other's private information, such as physical or electronic addresses, without explicit
  permission
* Other unethical or unprofessional conduct.

Project maintainers have the right and responsibility to remove, edit, or reject comments, commits,
code, wiki edits, issues, and other contributions that are not aligned to this Code of Conduct. By
adopting this Code of Conduct, project maintainers commit themselves to fairly and consistently
applying these principles to every aspect of managing this project. Project maintainers who do not
follow or enforce the Code of Conduct may be permanently removed from the project team.

This code of conduct applies both within project spaces and in public spaces when an individual is
representing the project or its community.

Instances of abusive, harassing, or otherwise unacceptable behavior may be reported by opening an
issue or contacting one or more of the project maintainers.

This Code of Conduct is adapted from the [Contributor Covenant](http://contributor-covenant.org),
version 1.2.0, available at
[http://contributor-covenant.org/version/1/2/0/](http://contributor-covenant.org/version/1/2/0/)