# TODO
## Refactor
- Integrate with GHA for build/test
  - add release pipeline
  - Dynamic Changelog (in releases?)
- Increase test coverage
- Implement integration tests
- Implement proper server wait, not select{}
- Store md5 in session, avoiding duplicate calc
- New Makefile
- Propagate context

## Features
- Add exit codes, so CI could detect if scan failed
- False/positive ignore file
- Support for webserver when using localmode?
- Add fingerprint `<commitid>:<file>:<rule/signature>:<line>`
- ignore-repo (for github) - when providing only user or org but want to exclude certain repositories from being scanned
- Create GHA to use the tool
  - eat our own dog food - rvsecret scans rvsecret
- Build docker image
- Add statistics for updated/downloaded signatures how many are found and how many are loaded - currently 60 out of 21+42(63)
- Test signatures before importing

## Bugs
- C++ generates false-positives (cisco/mlspp repository) --> should try with different signatures perhaps (like gitleaks?)
- "View commit on gitlab" (webserver) when using `local-git-repo`
- Snyk report!
- failed cloning repository https://github.com/Senzing/knowledge-base.git, empty git-upload-pack given -> does not result in the tool exiting with 1
