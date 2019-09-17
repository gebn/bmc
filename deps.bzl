def _maybe(repo_rule, name, **kwargs):
    if name not in native.existing_rules():
        repo_rule(name = name, **kwargs)

def _kingpin(go_repository):
    _maybe(
        go_repository,
        name = "com_github_alecthomas_kingpin",
        importpath = "github.com/alecthomas/kingpin",
        tag = "v2.2.6",
    )

    _maybe(
        go_repository,
        name = "com_github_alecthomas_units",
        commit = "680d30ca31172657fa50e996eb82d790d1d8b96e",
        importpath = "github.com/alecthomas/units",
    )

    _maybe(
        go_repository,
        name = "com_github_alecthomas_template",
        commit = "fb15b899a75114aa79cc930e33c46b577cc664b1",
        importpath = "github.com/alecthomas/template",
    )

def _prometheus(go_repository):
    _maybe(
        go_repository,
        name = "com_github_prometheus_client_golang",
        importpath = "github.com/prometheus/client_golang",
        tag = "v1.1.0",
    )

    _maybe(
        go_repository,
        name = "com_github_prometheus_common",
        importpath = "github.com/prometheus/common",
        tag = "v0.7.0",
    )

    _maybe(
        go_repository,
        name = "com_github_beorn7_perks",
        importpath = "github.com/beorn7/perks",
        tag = "v1.0.1",
    )

    _maybe(
        go_repository,
        name = "com_github_prometheus_client_model",
        importpath = "github.com/prometheus/client_model",
        commit = "14fe0d1b01d4d5fc031dd4bec1823bd3ebbe8016",
    )

    _maybe(
        go_repository,
        name = "com_github_prometheus_procfs",
        importpath = "github.com/prometheus/procfs",
        tag = "v0.0.5",
    )

    _maybe(
        go_repository,
        name = "com_github_matttproud_golang_protobuf_extensions",
        importpath = "github.com/matttproud/golang_protobuf_extensions",
        commit = "c182affec369e30f25d3eb8cd8a478dee585ae7d",
    )

def deps(go_repository):
    _maybe(
        go_repository,
        name = "com_github_google_gopacket",
        importpath = "github.com/google/gopacket",
        commit = "2d7fab0d91d6bb77c1df6bdeb222270a2aa13820",
    )

    _maybe(
        go_repository,
        name = "com_github_cenkalti_backoff",
        importpath = "github.com/cenkalti/backoff",
        tag = "v3.0.0",
    )

    _kingpin(go_repository)
    _prometheus(go_repository)

def test(go_repository):
    _maybe(
        go_repository,
        name = "com_github_google_go_cmp",
        importpath = "github.com/google/go-cmp",
        tag = "v0.3.1",
    )
