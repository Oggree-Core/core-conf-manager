
We need to set 3 env variable

REPO_URL

    REPO_URL=git@github.com:Oggree-core/core-env.git

GITHUB_KEY
SSH key file as base64 encoded string

GITHUB_KEY_PUB
SSH key.pub file as base64 encoded string


Target repo must have configs.json in root directory
    
    [
        {
            "target_folder": "caddy-config",
            "source_file": "Caddyfile",
            "target_file": "Caddyfile"
        },
        {
            "target_folder": "headscale-config",
            "source_file": "HeadscaleConfig.yaml",
            "target_file": "config.yaml"
        },
        {
            "target_folder": "headplane-config",
            "source_file": "HeadplaneConfig.yaml",
            "target_file": "config.yaml"
        }
    ]

Original Config Files must be inside the configs folder