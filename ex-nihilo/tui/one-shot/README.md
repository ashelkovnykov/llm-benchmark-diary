## Prompt

The following prompt was used for all of the models. It took ~35 minutes to write & edit.

> I'm starting a new personal project and I want your help writing it. It's going to be a Golang project that uses a TUI to interactively navigate a SQL DB. To start, here's what the skeleton for the project needs to be able to do:
>
> 1. The project should use the `bubbletea` Golang TUI packages
> 2. The project should use SQLite as the DB
> 3. Upon launching the process, it should list the existing DBs available to load in the `./db/` directory, and prompt the user to choose one. This directory may be empty at first.
> 4. While being prompted, the user should have the option to generate a new DB. This should prompt the user to choose a name for the new DB. This name should be validated: no empty strings, no reusing the name of an existing DB, and only alphanumeric characters. The user should be notified if the name is invalid and for what reason. If the name is valid, the DB should be saved with the others in `./db/`.
> 5. The DB should be generated using a schema that lives in `./internal/db/schema.sql`. You don't need to know anything about the real DB schema that will be used, other than that it's guaranteed to be valid, that it contains a table named "records", and that it uses a column named "rowid" as the primary key. You should just use a placeholder schema for now that meets these requirements. If you want to know more about the real DB schema for some reason, ask me.
> 6. If the user generates a new DB, it should be immediately loaded
> 7. Once a DB is loaded, the TUI should go to another screen. On this screen, the data contained in the DB will be visualized. Later, this will be an advanced and interactive process. For now, you should just generate a simple visualization: the total number of records contained in the DB, and up to the first 10 rows of data. Just enough to confirm that the table was loaded correctly.
> 8. Eventually, I would like to be able to import data into the DB via the TUI. For now, though, when a new table is generated it should be populated with 15 dummy records.
> 9. You can choose whatever name you want for the project / module. I'll change it myself later if necessary.
> 10. The binary for the compiled project should be output to `./bin/`.
>
> Feel free to create whatever files you think are necessary. However, at a minimum, I expect the following files to exist in this Go project:
>
> 1. `Makefile`: a makefile for performing typical Go project actions: formatting, building, running the binary, cleaning up files, etc.
> 2. `go.mod`: should include the dependencies necessary to build & run the project
> 3. `./cmd/main.go`: the entry point used to start the process
> 4. `./internal/db/schema.sql`: placeholder schema
>
> In addition, I expect `./internal/` to contain helper files which define the core logic for the project, in one or more modules.
>
> Be concise. You don't need to explain what you're doing or why you're doing it. I just need the source files for the project. If I have a question about something, I'll ask you myself. If I have an adjustment that I want you to make, I'll let you know. Ask clarifying questions before writing anything if you think that something is ambiguous / underspecified. If you have suggestions to make, list them _after_ providing the initial output as requested.

## Results

### Alibaba

#### Qwen 3.7 Plus

**Release Date:** June 2, 2026 \
**Time:** 2m 10s \
**Cost:** 5 credits \
**App Quality:** 7/10 \
**Code Quality:** 15.5/20

**Notes:**

- Process:
    - Qwen 3.7 Plus didn't label any of its files, just dumped them in unlabelled code blocks
- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot (this is okay because DBs can't be unloaded)
    - Only fetches DB records on DB load
    - Embeds schema into code
    - Initializes `./db/` directory on boot
    - Reuses a single `regexp` struct for new DB name validation
    - Uses custom `Record` struct for DB records
    - Uses `strings.Builder` for rendering
    - Compiles without error
- Cons:
    - Doesn't use a constant for the `.db` extension
    - Identical dummy records for each DB
    - No unit tests
    - Non-idiomatic import blocks
    - Some repetition in the update helpers that could be removed by using existing Bubble Tea bubbles or rewriting `Model.Update()`

#### Qwen 3.7 Max

**Release Date:** May 20, 2026 \
**Time:** 1m 29s \
**Cost:** 26 credits \
**App Quality:** 9/10 \
**Code Quality:** 15/20

**Notes:**

- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot, and appends any new DBs created
    - Only fetches DB records on DB load
    - Embeds schema into code
    - Initializes `./db/` directory on boot and before writing new DB
    - Reuses a single `regexp` struct for new DB name validation
    - Uses a custom `Record` struct for returning fetched DB records
    - Uses `strings.Builder` for rendering
    - Compiles without error
- Cons:
    - Schema embedded as `embed.FS` instead of `[]byte` / `string`
    - Doesn't use a constant for the `.db` extension
    - Identical dummy records for each DB
    - No unit tests
    - Non-idiomatic import blocks
    - `os.IsNotExist` in `ListDBs` should be fatal, since `EnsureDBDir` should have been called by then

### Anthropic

#### Claude Haiku 4.5

**Release Date:** October 1, 2025 \
**Time:** 12s \
**Cost:** 12 credits \
**App Quality:** 5/10 \
**Code Quality:** 9.5/20

**Notes:**

- Process:
    - Haiku 4.5 had follow-up questions, documented in [RESPONSE.md](./haiku-4.5/RESPONSE.md)
    - Haiku 4.5 didn't label any of its files, just dumped them in unlabelled code blocks
- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot and on new DB name validation (this is okay because DBs can't be unloaded)
    - Only fetches DB records on DB load
    - Initializes `./db/` directory on boot
    - Uses `strings.Builder` for rendering
- Cons:
    - Doesn't embed the schema into the code
    - Doesn't use a constant for the `.db` extension
    - Doesn't reuse `regexp` struct
    - Doesn't return DB records using a custom struct; renders them directly to `string`
    - Identical dummy records for each DB
    - No unit tests
    - Non-idiomatic import blocks
    - Tries to render integers to runes using ASCII match, but it doesn't work correctly at all
        - This breaks the `tui.toString()` helper, which isn't even used in every case where it tries to do ASCII math
    - Rendering the DB records in-place without helpers is ugly and dumb
    - Crashes on new DB generation without these changes
        - `Model.createInput` needs to have type `*strings.Builder{}`
        - `Model.createInput` needs to be initialized in `NewInitialModel()`

#### Claude Opus 4.8

**Release Date:** May 27, 2026 \
**Time:** 37s \
**Cost:** 66 credits \
**App Quality:** 8/10 \
**Code Quality:** 14/20

**Notes:**

- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` before navigating to the DB selection screen
    - Only fetches DB records on DB load
    - Embeds schema into code
    - Initializes `./db/` directory before reading contents or writing new DB
    - Reuses a single `regexp` struct for new DB name validation
    - Uses custom `Snapshot` & `Row` structs for returning fetched DB records
    - Uses `strings.Builder` for rendering
    - Compiles without error
- Cons:
    - Doesn't use a constant for the `.db` extension
    - Custom `Row` struct renders contents of each column to `string`
    - Identical dummy records for each DB
    - No unit tests
    - Import blocks out of order
    - Mixes value and pointer receivers
    - Could stand to split up `tui.go` into several files

### Google

#### Gemini 3.1 Pro

**Release Date:** December 31, 2024 \
**Time:** 34s \
**Cost:** 18 credits \
**App Quality:** 8/10 \
**Code Quality:** 12.5/20

**Notes:**

- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot and before navigating to the DB selection screen
    - Only fetches DB records on DB load
    - Uses a custom `Record` struct for returning fetched DB records
    - Compiles without error
- Cons:
    - Doesn't embed the schema into the code
    - Doesn't reuse `regexp` struct for name validation
    - Doesn't use a `strings.Builder` for rendering
    - Doesn't use a constant for the `.db` extension
    - `./db/` directory management is all over the place: `Makefile`, `main.go`, `db.go`
    - Identical dummy records for each DB
    - No unit tests
    - Non-idiomatic import blocks

### MiniMax

#### MiniMax M2.7

**Release Date:** March 17, 2026 \
**Time:** 1m 4s + ~15m \
**Cost:** 4 credits \
**App Quality:** 6/10 \
**Code Quality:** 6.5/20

**Notes:**

- Process:
    - MiniMax M2.7 didn't label any of its files, just dumped them in unlabelled code blocks
    - MiniMax M2.7 seemed to get very confused, generating multiple copies of some files:
        - 3 versions of the `main` function
        - 3 versions of the `New` function
        - 2 versions of `dbListModel` and methods
        - 5 versions of `dbViewModel` and methods
- Pros:
    - Only fetches the contents of `./db/` on boot and on new DB name validation (this is okay because DBs can't be unloaded)
    - Only fetches DB records on DB load
    - Uses a simple ASCII range scan for DB name validation
    - Uses a custom `Record` struct for returning fetched DB records
    - Initializes `./db/` directory before writing new DB
- Cons:
    - `schema.sql` is unused; instead, the schema is hard-coded into `schema.go`
    - Doesn't use a constant for the `.db` extension
    - Doesn't use a `strings.Builder` for rendering
    - Identical dummy records for each DB
    - No unit tests
    - Uses an odd two-phase "load-then-view" solution in `main.go` that's difficult to extend without rewrites
    - Non-idiomatic import blocks
    - `models.go` is broken and unused
    - `input.go` is unused
    - `dbListModel.Update()` is needlessly over complicated and contains a bug that prevents backspace from being used when entering in a new DB name
    - `onCancel` in `dbListModel` does nothing
    - Errors returned as `string` instead of `error` during name validation
    - App does not compile without significant manual editing of output code (20+ minutes of clean-up). To the LLM's credit, though, no new code needs to be added.

#### MiniMax M3

**Release Date:** May 30, 2026 \
**Time:** 3m 38s \
**Cost:** 17 credits \
**App Quality:** 7/10 \
**Code Quality:** 15.5/20

**Notes:**

- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot and if DB fails to load (this is okay since DBs can't be unloaded)
    - Only fetches DB records on DB load
    - Embeds schema into code
    - Initializes `./db/` directory before reading from / writing to it
    - Reuses a single `regexp` struct for new DB name validation
    - Uses custom `Record` struct for DB records
    - Uses `strings.Builder` for rendering
    - Compiles without error
- Cons:
    - Doesn't use a constant for the `.db` extension
    - Helpers, methods, and structs all jumbled together in `tui.go`
    - Could stand to split up `tui.go` into several files
