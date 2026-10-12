## Summary

This repo contains experiments comparing various LLM models, tools, and worfkows. Each experiment has an explicit goal, whether that's fixing a bug, implementing a feature, or creating a product. Both the output and process of accomplishing the goal are recorded, so that the models, tools, and workflows can be directly compared against each other with regards to correctness, speed, quality, ease, and cost.

## Motivation

Since 2023, I've naturally ossified into my preferred LLM workflow without ever trying too many alternative flows. It seems that all of my friends and coworkers have, too. Which workflows work best under different circumstances is still an open and hotly debated topic. I want to force myself to try a breadth of LLM programming tools in order to identify the relative strengths and weaknesses of each. I want to do this in public, in case others can derive utility from this process, and in order to have any mistakes I make along the way be corrected. I'd like to identify the optimal settings for each type of LLM workflow, as well as the optimal conditions under which to switch between them. However, any experience that will allow me to incrementally improve my workflow is still a win.

## Table of Contents

- **[ex-nihilo/](./ex-nihilo/README.md)**: Experiments that involve code generation from scratch with no existing codebase
  - **[tui/](./ex-nihilo/tui/README.md)**: TUI DB explorer experiment
    - **[control/](./ex-nihilo/tui/control)**: 100% hand-written solution, comparison purposes
    - **[one-shot/](./ex-nihilo/tui/one-shot)**: The LLM is directly asked for a solution using a master prompt; no user feedback, follow-up prompts, debug cycles, or iterative corrections
