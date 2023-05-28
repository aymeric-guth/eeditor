## Description

A generic multi-host binding for EDITOR environment variable

## Config

### Lookup path

EEDITOR_CONFIG > XDG_CONFIG_HOME/eeditor/eeditor.yml > ~/.config/eeditor/eeditor.yml > ~/.eeditor.yml

### Example

```yaml
---
# sequential evaluation order
# no path to executable specified = PATH lookup ~which
- name: nvim
# one path is specified
- name: nvim
  path: $TOOLDIR/bin
# multiple path specified, evaluation order = definition order
- name: nvim
  path:
    - $TOOLDIR/bin
    - /usr/local/bin
# additional environment variables are passed at command invocation
- name: nvim
  env:
    XDG_CONFIG_HOME: $DOTFILES
# alternative editors if none of the above are available
- name: vim
  path: /usr/bin
- name: vi
  path: /usr/bin
```

## Installation

```shell
go install gitea.ars-virtualis.org/yul/eeditor@latest
```
