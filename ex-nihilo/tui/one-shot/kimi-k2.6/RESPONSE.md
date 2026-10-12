## Kimi K2.6

A few quick questions:

1. What file extension should database files use in `./db/`? (e.g., `.db`, `.sqlite`, all files?)
2. Besides `rowid`, what columns should the placeholder `records` table have? This affects the dummy data and the visualization.
3. Are you open to using `charmbracelet/bubbles` and `lipgloss` for inputs/tables, or do you want minimal dependencies (just `bubbletea` + driver)?

## Me

1. They should use `.db` as their file extension
2. Dummy columns like "name" and "value" will suffice
3. Using existing bubbles is perfectly acceptable, if you consider that to be the best choice for the implementation
