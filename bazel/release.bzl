load("@io_bazel_rules_docker//go:image.bzl", "go_image")
load("@io_bazel_rules_docker//container:container.bzl", "container_image")
load("@io_bazel_rules_docker//container:push.bzl", "container_push")

def release(target):
    service_name = target.replace("//cmd/", "")

    real_target = target + ":" + service_name + "_lib"

    container_image(
        name = service_name + "_container_image",
        base = "@1.19.5-bullseye//image",
        entrypoint = ["/" + target.replace("//cmd/", "")],
        files = [target + ":" + target.replace("//cmd/", "")],
    )
    container_push(
        name = "release-" + service_name,
        image = service_name + "_container_image",
        format = "Docker",
        registry = "ghcr.io",
        repository = "openchainxyz/" + service_name,
        tag = "{" + service_name + "-version}",
    )