# rvsecret

<img src="https://raw.githubusercontent.com/MariaLetta/free-gophers-pack/master/illustrations/svg/2.svg" width=200 height=200>

A hard fork of [wraith](https://github.com/N0MoreSecr3ts/wraith) (which itself is a hard fork of [gitrob](https://github.com/michenriksen/gitrob))

![build status](https://github.com/rumenvasilev/rvsecret/actions/workflows/on-push.yaml/badge.svg)
[![codecov](https://codecov.io/gh/rumenvasilev/rvsecret/graph/badge.svg?token=X2BXUU5H0S)](https://codecov.io/gh/rumenvasilev/rvsecret)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=rumenvasilev_rvsecret&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=rumenvasilev_rvsecret)

rvsecret detects secrets in git repositories (or localpath). The big differentiator is it is tightly integrated with Github and Gitlab APIs and has no extra dependencies. It's just one binary you can use anywhere (even in scratch containers).

## It is still very much WIP!

The tool is undergoing a big refactor at the moment, hence it's not stable to be used yet. For detailed refactoring plans checkout [TODO.md](TODO.md).

## Capabilities

### Targets
- Gitlab.com repositories and projects
- Github.com repositories and organizations
- Local git repositories
- Local filesystem

### Major Features

- Exclude files, paths, and extensions
- Web and terminal interfaces for real-time results (very much alpha)
- Configurable commit depth

## Screenshots
<p>
  <img width="537" alt="Screen Shot 2020-08-16 at 11 23 25 PM" src="https://user-images.githubusercontent.com/672940/90354541-9f515a80-e017-11ea-8669-97a2d7823cbb.png">
  <img width="365" alt="Screen Shot 2020-08-16 at 11 23 43 PM" src="https://user-images.githubusercontent.com/672940/90354550-a11b1e00-e017-11ea-9bb6-5f7c6209f7b0.png">
</p>
<br>

## Quickstart

1. Compile it (`make build`)
2. Download the [signature file](https://github.com/rumenvasilev/wraith-signatures/blob/main/signatures/default.yaml) to `~/.rvsecret/signatures/`
3. Use it:

```bash
bin/rvsecret-darwin scan local-git-repo --local-repos rumenvasilev/terraform
```

## Documentation

### Build from source
To build from source, you need Go 1.21.
```shell
    $ cd $GOPATH/src
    $ git clone git@github.com:rumenvasilev/rvsecret.git
    $ cd rvsecret
    $ make build
    $ ./bin/rvsecret-<ARCH> <sub-command>
```

### Signatures
Signatures are the current method used to detect secrets within a target source. They are broken out into the [wraith-signatures][4] repo for extensability purposes. This allows them to be independently versioned and developed without having to recompile the code. To make changes just edit an existing signature or create a new one. Check the [README][5] in that repo for additional details.

### Authencation
rvsecret will need either a GitLab or Github access token in order to interact with their appropriate API's. You can create a [GitLab personal access token][6], or [a Github personal access token][7] and save it in an environment variable in your **bashrc**, add it to a rvsecret config file, or pass it in on the command line. Passing it in on the commandline should be avoided if possible for security reasons. Of course if you want to eat your own dog food, go ahead and do it that way, then point rvsecret at your command history file. :smiling_imp:

## Contributing

Fork, PR and lets have a chat. Outstanding work is in [TODO.md](TODO.md).

[3]: https://github.com/rumenvasilev/rvsecret/releases
[4]: https://github.com/rumenvasilev/wraith-signatures
[5]: https://github.com/rumenvasilev/wraith-signatures/blob/master/README.md
[6]: https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html
[7]: https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/

[12]: https://github.com/dxa4481/truffleHog
