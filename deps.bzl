load("@bazel_gazelle//:deps.bzl", "go_repository")

def _maybe(repo_rule, name, **kwargs):
    if name not in native.existing_rules():
        repo_rule(name = name, **kwargs)

def _kingpin():
    _maybe(
        go_repository,
        name = "com_github_alecthomas_kingpin",
        importpath = "github.com/alecthomas/kingpin",
        tag = "v2.2.6",
    )

    _maybe(
        go_repository,
        name = "com_github_alecthomas_units",
        commit = "2efee857e7cfd4f3d0138cc3cbb1b4966962b93a",  # master as of 2015-10-22
        importpath = "github.com/alecthomas/units",
    )

    _maybe(
        go_repository,
        name = "com_github_alecthomas_template",
        commit = "a0175ee3bccc567396460bf5acd36800cb10c49c",  # master as of 2016-04-05
        importpath = "github.com/alecthomas/template",
    )

def deps():
    _maybe(
        go_repository,
        name = "com_github_google_gopacket",
        importpath = "github.com/google/gopacket",
        commit = "7cc6592eca24b42f05b6d13a5521d9d2558cf53b",
    )

    _maybe(
        go_repository,
        name = "com_github_cenkalti_backoff",
        importpath = "github.com/cenkalti/backoff",
        tag = "v3.0.0",
    )

    _kingpin()

def test():
    _maybe(
        go_repository,
        name = "com_github_google_go_cmp",
        importpath = "github.com/google/go-cmp",
        tag = "v0.3.0",
    )
