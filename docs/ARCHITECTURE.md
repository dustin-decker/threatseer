# Threatseer Architecture

<p align="center">
  <img src="images/components.svg" width="100%"/>
</p>

## Analysis engines

The server's analysis engines each contribute a risk score to an event as it flows through the pipeline. Actions can be taken depending on the overall score.

<p align="center">
  <img src="images/engine_pipeline.svg" width="100%"/>
</p>

### Static Analysis Engine

The Static Analysis Engine has configurable checks such as risky proceses, file/directory integrity monitoring (todo), and known IOCs (todo).

### Dynamic Rules Engine

The Dynamic Rules Engine allows for the user to prove custom query-based rules that will apply a score and an action to the events matched.

### Profile Engine

Automatically generates execution profile for the best reliable identifier available (container id or process id right now). Applies a positive or negative score to the event depending on if the behavior is outside, or inside the profile.

### Classification Engine (todo)

Machine learning prediction model trained on two sets of labeled data:

#### GOOD

- A baseline healthy environment

#### BAD

- ATT&CK test automation
- Red team activity
- Public exploits
- Honeypots
- Events triaged by SOC

### Action Engine

Takes action depending on final pipeline score. 
Actions are `toss`, `log`, `alert`, and `kill`. (todo)
