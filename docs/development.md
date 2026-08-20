# Development

Use the repo-local validation entrypoint:

```sh
bin/container bin/ci
```

If container support is not ready for this repo yet:

```sh
bin/ci
```

Document language runtimes, package managers, environment variables, and local
service dependencies here as the project becomes concrete.

When host-local language execution is supported, commit exact versions in a
root `mise.toml` and install them with:

```sh
mise install
```

Container-only repositories may omit host runtime pins. Application packages
remain owned by the ecosystem manifest and lockfile.
