# TODO
## Refactor
- x Remove dependency on bindata - github.com/elazarl/go-bindata-assetfs
- x Change banner
- x Remove search for all github orgs if none is provided
- Integrate with GHA for build/test
  - x build,test pipeline
  - x unit test, coverage (https://codecov.io/bash?), race
  - add release pipeline
  - Dynamic Changelog (in releases?)
- Increase test coverage
- Implement integration tests
- Refactor core package - remove multi-level nesting
- Remove dependencies on Hashicorp as much as possible (viper?)
- Create config package
- Remove global cfg var, explicitly pass it as arg
- x os.Exit() everywhere, raise error instead
- replace pkg.Scan with interface for easier testing
- Implement proper server wait, not select{}
- Store md5 in session, avoiding duplicate calc
- New Makefile
- Overhaul updateRules.go
- x Start webserver everywhere

## Features
- Add exit codes, so CI could detect if scan failed
- x Commit depth configurable
- False/positive ignore file
- Support for webserver when using localmode?
- Add fingerprint `<commitid>:<file>:<rule/signature>:<line>`
- ignore-repo (for github) - when providing only user or org but want to exclude certain repositories from being scanned
- Create GHA to use the tool
  - eat our own dog food - rvsecret scans rvsecret
- Build docker image

## Bugs
- C++ generates false-positives (cisco/mlspp repository) --> should try with different signatures perhaps (like gitleaks?)
- "View commit on gitlab" (webserver) when using `local-git-repo`
- Snyk report!
- failed cloning repository https://github.com/Senzing/knowledge-base.git, empty git-upload-pack given -> does not result in the tool exiting with 1
