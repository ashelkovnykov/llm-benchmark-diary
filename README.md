## Summary

This repo contains several experiments for comparing LLM models and worfkows. Each experiment has an explicit goal, whether that's solving a problem or creating a product. The output and process of accomplishing the goal of each experiment is recorded for a variety of models and workflows. The purpose of these experiments is to direclty compare various models and workflows for correctness, speed, quality, ease, and cost.

## Motivation

Since 2023, I've naturally ossified into my preferred LLM workflow without ever trying too many alternative flows. It seemed to me that all of my friends and coworkers have, too. I want to force myself to try a breadth of LLM models and worflows in order to identify the relative strengths and weaknesses of each. I want to identify which models and tools work best for which tasks and workflow preferences. I want to do this in public, in case others can derive utility from this process, and in order to have any mistakes I make along the way be corrected.

## Table of Contents

- **[ex-nihilo/](./ex-nihilo/README.md)**: Experiments that involve code generation from scratch with no existing codebase
  - **[tui/](./ex-nihilo/tui/README.md)**: TUI DB explorer experiment
    - **[control/](./ex-nihilo/tui/control)**: 100% hand-written solution, comparison purposes
    - **[one-shot/](./ex-nihilo/tui/one-shot)**: The LLM is directly asked for a solution using a master prompt; no user feedback, follow-up prompts, debug cycles, or iterative corrections
