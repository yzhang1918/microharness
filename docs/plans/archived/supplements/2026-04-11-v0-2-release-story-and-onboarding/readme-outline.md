# README Outline Draft

This supplement captures the intended README shape for the `0.2.0` release so
the execution phase can rewrite the root README without depending on discovery
chat.

## Hero Direction

- Primary line: `Harnesses matter. Building one shouldn't be the project.`
- Short follow-up: `easyharness` is a git-native, agent-first harness for
  human-steered, agent-executed work.
- Longer follow-up: `easyharness` packages plans, repo-local instructions,
  workflow state, review, and archive paths into a thin system that coding
  agents can actually follow, so humans can steer work without micromanaging
  every implementation detail.

## README Priorities

The README should help a human engineer or team answer these questions in
order:

1. What is `easyharness`?
2. Why does it exist?
3. How do I try it quickly?
4. What should the human do versus the agent?
5. Where do I go for deeper setup, workflow, and specs?

## Suggested Section Order

1. Hero and one-paragraph positioning
2. Why `easyharness`
3. Quickstart
4. Stability
5. How humans steer
6. Further reading
7. Contributor note linking to `docs/development.md`
8. Deeper product/workflow/spec links

## Why `easyharness`

The README should explain the product in terms of the problem it solves, not
only the repository layout. Good points to carry:

- long-running agent work loses coherence when plans, state, and rules live
  only in chat history
- ad hoc shell scripts and prompt glue create hidden workflow that is hard to
  inspect, review, or teach to the next agent run
- humans should spend attention on direction, approval, and execution summaries
  rather than line-by-line micromanagement

## Quickstart Draft

Keep this section short:

```bash
brew install easyharness
cd /path/to/your-repo
harness init
```

Then explain in one or two short paragraphs:

- `harness init` installs the repo-local instructions and skills that tell the
  agent how to work in that repository
- after init, restart the coding agent so the new `AGENTS.md` block and
  `.agents/skills/` are picked up cleanly

## Stability Draft

Use a durable statement rather than a one-time launch note.

Recommended content:

- `easyharness` is evolving quickly, and breaking changes may happen between
  releases
- that does not mean the human operator must manually track every internal
  detail; the harness is designed so agents can recover the relevant workflow
  context from repo-local instructions, plans, state, and skills
- current evolution should continue to favor lower agent cognitive load,
  stronger execution quality, and better human steering without micromanage

## How Humans Steer

This section should explain the human role in practical terms:

- set intent, scope, and constraints
- approve or adjust plans
- review execution summaries, outcomes, and high-signal artifacts
- step in where product or risk judgment is required
- avoid treating success as reviewing every changed line manually

## Further Reading

The README may link to these as supporting context:

- [Harness design for long-running apps](https://www.anthropic.com/engineering/harness-design-long-running-apps)
- [Harness engineering: leveraging Codex in an agent-first world](https://openai.com/index/harness-engineering/)

These should support the explanation, not replace it.

## Contributor Note

The contributor/development pointer should be short:

- this repository is developed primarily through agents
- repo-specific operating guidance lives in `AGENTS.md`
- detailed local development setup and maintainer workflows live in
  `docs/development.md`
