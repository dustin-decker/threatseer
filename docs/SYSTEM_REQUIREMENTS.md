# Threatseer System Requirements

## Agent

The sensor requires kernel support for KProbes and Perf

### Minimum Distro/Kernel Requirements

- Linux kernel 3.10+ 
- x86_64 (possibly x86_32), but not ARM

There are some exceptions due to maintainers backporting kernel features in some instances.

Major distribution versions with minimum requirements:

    - Ubuntu 14.04.0 (kernel version 3.13)
    - Debian 8 (kernel version 3.16)
    - Fedora 21 (kernel version 3.17)
    - CentOS 6.6 (kernel version 2.6.32-504)
    - CentOS 7 (with kernel version 3.10)

### Recommended Requirements
- Linux kernel 4.4+

Major distribution versions with recommended requirements:

    - Ubuntu 14.04.5+ (kernel version 4.4)
    - Debian 9+ (kernel version 4.9)
    - Fedora 24+ (kernel version 4.5)

### Resource Requirements

- Typically less than ~1% of one CPU
- ~30 MB RAM
- ~15 MB disk space

## Server

Server resource consumption depends on how many agents are connected and how much telemetry the agents are pushing.
This will likely change over time as more behavioral analysis engines and rules are added.
Resource consumption starting from:

- Typically less than ~1% of one CPU
- ~10 MB RAM
- ~30 MB disk space