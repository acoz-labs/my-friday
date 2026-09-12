#!/bin/sh
# Keep native shell setup from selecting another installed My Friday version.
# The optional override is machine-local, never stored in the memory bank.
friday_memory_binary=${MY_FRIDAY_MEMORY_BIN:-my-friday}
case ${MY_FRIDAY_MEMORY_BIN:-} in
  ''|/*) ;;
  *)
    if [ "$1" = hook ]; then
      printf '%s\n' '{"systemMessage":"My Friday memory runtime override must be an absolute executable path; no memory was changed."}'
      exit 0
    fi
    printf '%s\n' 'My Friday memory runtime override must be an absolute executable path.' >&2
    exit 1
    ;;
esac

case "$1" in
  hook)
    # A missing/older runtime must not turn optional recall into a blocking hook.
    # Suppress failed raw output, which need not be valid JSON or safe context.
    if friday_memory_output=$("$friday_memory_binary" codex-memory-hook 2>/dev/null); then
      printf '%s\n' "$friday_memory_output"
    else
      printf '%s\n' '{"systemMessage":"My Friday memory hook is unavailable. Check MY_FRIDAY_MEMORY_BIN, the installed version and the memory binding; automatic recall did not run."}'
    fi
    ;;
  mcp)
    exec "$friday_memory_binary" mcp --harness codex
    ;;
  *) exit 1 ;;
esac
