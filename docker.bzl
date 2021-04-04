CONFIG_SETTING_NAME = "docker_go_config"

PLATFORMS_RULE = "@io_bazel_rules_go//go/toolchain:linux_amd64"

def register_go_config_setting():
    native.config_setting(
        name = CONFIG_SETTING_NAME,
        values = {"platforms": PLATFORMS_RULE},
        visibility = ["//visibility:private"],
    )

def select_go_base():
    # Use select to make sure that we don't accidentally build non-linux binaries.
    return select(
        {
            ":" + CONFIG_SETTING_NAME: "@go_image_base//image",
        },
        no_match_error = "You must specify --config=linux_amd64 for docker rules",
    )
