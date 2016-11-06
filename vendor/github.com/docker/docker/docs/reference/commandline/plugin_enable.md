---
title: "plugin enable"
description: "the plugin enable command description and usage"
keywords: "plugin, enable"
advisory: "experimental"
---

<!-- This file is maintained within the docker/docker Github
     repository at https://github.com/docker/docker/. Make all
     pull requests against that repo. If you see this file in
     another repository, consider it read-only there, as it will
     periodically be overwritten by the definitive file. Pull
     requests which include edits to this file in other repositories
     will be rejected.
-->

# plugin enable (experimental)

```markdown
Usage:  docker plugin enable PLUGIN

Enable a plugin

Options:
      --help   Print usage
```

Enables a plugin. The plugin must be installed before it can be enabled,
see [`docker plugin install`](plugin_install.md).


The following example shows that the `no-remove` plugin is installed,
but disabled:

```bash
$ docker plugin ls

NAME                  TAG                 DESCRIPTION                ENABLED
tiborvass/no-remove   latest              A test plugin for Docker   false
```

To enable the plugin, use the following command:

```bash
$ docker plugin enable tiborvass/no-remove

tiborvass/no-remove

$ docker plugin ls

NAME                  TAG                 DESCRIPTION                ENABLED
tiborvass/no-remove   latest              A test plugin for Docker   true
```

## Related information

* [plugin ls](plugin_ls.md)
* [plugin disable](plugin_disable.md)
* [plugin inspect](plugin_inspect.md)
* [plugin install](plugin_install.md)
* [plugin rm](plugin_rm.md)
