# 0.3.0

BACKWARDS INCOMPATIBILITIES:

FEATURES:

- added mutual TLS support ([#15](https://github.com/dustin-decker/threatseer/pull/15))
    - enabled configurable server endpoint for agent

IMPROVEMENTS:

- exposed some Profile Engine tunables 
- use LRU cache for tracking ongoing execution profiling ([#12](https://github.com/dustin-decker/threatseer/issues/12))

BUG FIXES:

None

# 0.2.0

BACKWARDS INCOMPATIBILITIES:

- threatseer config changed

FEATURES:

- added Profile Engine for automatic executable and container image execution profiling

IMPROVEMENTS:

- cache Dynamic Engine rule ASTs
- buffer events for engine pipeline
- JSON logging improvements

BUG FIXES:

None


# 0.1.1

BACKWARDS INCOMPATIBILITIES:

None

FEATURES:

None

IMPROVEMENTS:

None

BUG FIXES:

- vendored various upstream fixes since capsule8 0.12.0 release

# 0.1.0

initial release