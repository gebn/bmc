load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "e5de048e72612598c45f564202f6a3c74616be4ffd2dbd6f7bc75045f8ecbdce",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.23.4/rules_go-v0.23.4.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.23.4/rules_go-v0.23.4.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains()

http_archive(
    name = "bazel_gazelle",
    sha256 = "cdb02a887a7187ea4d5a27452311a75ed8637379a1287d8eeb952138ea485f7d",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.21.1/bazel-gazelle-v0.21.1.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.21.1/bazel-gazelle-v0.21.1.tar.gz",
    ],
)

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

gazelle_dependencies()

http_archive(
    name = "com_google_protobuf",
    sha256 = "e5265d552e12c1f39c72842fa91d84941726026fa056d914ea6a25cd58d7bbf8",
    strip_prefix = "protobuf-3.12.3",
    urls = ["https://github.com/protocolbuffers/protobuf/archive/v3.12.3.zip"],
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

load(":deps.bzl", "deps", "test")

deps(go_repository)

test(go_repository)
