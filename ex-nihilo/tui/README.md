## Summary

This experiment tasks the LLM with creating a simple TUI for browsing the contents of a SQL DB. The project is described to the LLM as eventually being much larger, with advanced interactive visualizations and the ability to live edit DBs through the TUI. However, the LLM is at first tasked only with generating a "v1" or "skeleton" for the project. The expected output is a TUI written in Golang using the Bubble Tea library that can load a SQLite DB stored locally and visualize the total number of columns in the DB, as well as the first 10 rows. It also needs to be able to generate new DBs and seed them with dummy data for visualization.

## Results

### One-shot

| Model             | Active Time | Passive Time | Total Time | Cost | App Quality | Code Quality |
|-------------------|-------------|--------------|------------|------|-------------|--------------|
| Control (None)    | ~10h        | 0s           | ~10h       | free | 6 / 10      | 14   / 20    |
| Qwen 3.7 Plus     | 35m         | 2m 10s       | 37m 10s    | 5    | 7 / 10      | 15.5 / 20    |
| Qwen 3.7 Max      | 35m         | 1m 29s       | 36m 29s    | 26   | 9 / 10      | 15   / 20    |
| Qwen 3.8 Flash    | 35m         | 13m 45s      | 48m 45s    | 11   | 8 / 10      | 13.5 / 20    |
| Qwen 3.8 Max      | 35m         | 1m 39s       | 36m 39s    | 15   | 7 / 10      | 15   / 20    |
| Claude Haiku 4.5  | ~37m        | 12s          | 37m 12s    | 12   | 5 / 10      |  9.5 / 20    |
| Claude Opus 4.8   | 35m         | 37s          | 35m 37s    | 66   | 8 / 10      | 14   / 20    |
| DeepSeek V4 Pro   | ~38m        | 3m 12s       | 41m 12s    | 8    | 9 / 10      |  9.5 / 20    |
| Gemini 3.1 Pro    | 35m         | 34s          | 35m 34s    | 18   | 8 / 10      | 12.5 / 20    |
| Muse Spark 1.2    | 35m         | 18s          | 35m 18s    | 8    | 4 / 10      | 12   / 20    |
| MiniMax M2.7      | ~50m        | 1m 4s        | ~51m 4s    | 4    | 6 / 10      |  6.5 / 20    |
| MiniMax M3        | 35m         | 3m 38s       | 38m 38s    | 17   | 7 / 10      | 15.5 / 20    |
| Kimi K2.6         | 35m         | 5m 14s       | 40m 14s    | 22   | 9 / 10      | 14   / 20    |
| Kimi K3           | 35m         | 49s          | 35m 49s    | 20   | 7 / 10      | 16   / 20    |
| GPT 5.6 Terra     | 35m         | 3m 8s        | 38m 8s     | 8    | 9 / 10      | 16   / 20    |
| GPT 5.6 Terra Pro | 35m         | 2m 47s       | 37m 47s    | 10   | 10 / 10     | 15.5 / 20    |
| GPT 5.6 Sol       | 35m         | 2m 13s       | 37m 13s    | 22   | 10 / 10     | 16   / 20    |
| GPT 5.6 Sol Pro   | 35m         | 1m 58s       | 36m 58s    | 28   | 9 / 10      | 15.5 / 20    |
| GLM 5.3 Flash     | 35m         | 11s          | 35m 11s    | 1    | 8 / 10      | 13   / 20    |

## Notes

1. The `Cost` column is in credits from my inference reseller. It correlates to the number of output tokens, albeit imperfectly since some models cost less / have higher output token limits.
2. `Active Time` is time that a user must spend actively working with the LLM (e.g. writing the prompt). `Passive Time` is time during which the LLM works independently and the user is free to pursue other tasks. `Total Time` is the sum of the two: the total wall-clock time required to complete the task.
3. `App Quality` is graded using a simple 10-point rubric:
    1. DB selection screen renders existing DBs correctly, handles empty list, free from bugs, etc.
    2. DB selection screen free from navigation bugs
    3. New DB naming screen handles keystrokes, renders error messages, free from bugs, etc.
    4. New DB naming screen rejects invalid names
    5. New DB is loaded immediately
    6. DB view screen includes both record count & row data
    7. DB view screen renders row data in a clear, neat, and column-aligned way
    8. DB view screen is free from rendering bugs
    9. Loaded DB capable of being unloaded and another DB reloaded
    10. Pleasing visual style using Lipgloss
4. `Code Quality` is graded with the following partially-subjective 20-point metric. See the notes for each particular experiment for how its rating was determined.
    1. (1 point) Manages the Bubble Tea screens using its own internal state machine instead of creating a graph of tea.Model structs
    2. (1 point) Only fetch `./db/` contents prior to rendering the selection screen (or at least don't refetch them on every re-render)
    3. (1 point) Only fetch the records from a loaded DB on first render
    4. (1 point) DB schema embedded into code
    5. (1 point) Includes a constant for the `.db` extension instead of manually pasting it everywhere
    6. (1 point) Independently manages the `./db/` directory (e.g. ensures creation, checks whether it's empty, etc.)
    7. (1 point) Reuses name validation regex struct, or uses an ASCII mask approach for name validation
    8. (1 point) Uses a custom struct for returning fetched DB records
    9. (1 point) Uses `strings.Builder` for rendering the TUI
    10. (1 point) Randomized, unique seed data for new DBs
    11. (2 points) Unit tests
    12. (3 points) Code style: package boundaries, helpers (or lack thereof), import block ordering, etc.
    13. (5 points) Code works without modification; -1 point for every bug that needs to be fixed to compile
