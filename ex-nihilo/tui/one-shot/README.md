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

#### Qwen 3.8 Flash

**Release Date:** August 25, 2026 \
**Time:** 13m 45s \
**Cost:** 11 credits \
**App Quality:** 8/10 \
**Code Quality:** 13.5/20

**Notes:**

- Process:
    - I wouldn't describe 14m response time as very "flash"
- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot and before navigating to the DB selection screen
    - Only fetches DB records on DB load and on demand
    - Embeds schema into code
    - Initializes `./db/` directory on boot and before reading from / writing to it
    - Reuses a single `regexp` struct for new DB name validation
    - Uses custom `Preview` struct for returning fetched DB records
    - Uses `strings.Builder` for rendering
    - Compiles without error
- Cons:
    - Doesn't use a constant for the `.db` extension
    - `Preview` struct holds record values as rendered `string`s
    - Identical dummy records for each DB
    - No unit tests
    - Parameterization in `Makefile` needs improvement
    - Loading the DB directory as an envvar is unnecessary given the project description, and more likely to cause bugs in the future
    - Needlessly complicated spaghetti helper logic in `tui.go`

#### Qwen 3.8 Max

**Release Date:** August 2, 2026 \
**Time:** 1m 39s \
**Cost:** 15 credits \
**App Quality:** 7/10 \
**Code Quality:** 15/20

**Notes:**

- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` before navigating to the DB selection screen
    - Only fetches DB records on DB load
    - Embeds schema into code
    - Reuses a single `regexp` struct for new DB name validation
    - Uses custom `Record` struct
    - Uses `strings.Builder` for rendering
    - Compiles without error
- Cons:
    - Doesn't use a constant for the `.db` extension
    - Schema embedded as `embed.FS` instead of `[]byte` / `string`
    - Never checks whether `./db/` exists, nor creates it if it doesn't. This leads to runtime errors if the directory doesn't exist.
    - Identical dummy records for each DB
    - No unit tests

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

#### Claude Fable 5.1

**Release Date:** August 31, 2026 \
**Time:** 1m 7s \
**Cost:** 182 credits \
**App Quality:** 9/10 \
**Code Quality:** 15.5/20

**Notes:**

- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot, before navigating to the DB selection screen, and before & after writing to it
    - Only fetches DB records on DB load
    - Embeds schema into code
    - Uses a constant for the `.db` extension
    - Initializes `./db/` directory on boot
    - Reuses a single `regexp` struct for new DB name validation
    - Uses `strings.Builder` for rendering
    - Compiles without error
- Cons:
    - Doesn't return DB records using a custom struct; renders them directly to `string`
    - Identical dummy records for each DB
    - No unit tests
    - Mixes value and pointer receivers
    - Could stand to split up `model.go` into several files

### DeepSeek

#### DeepSeek V4 Pro

**Release Date:** August 13, 2026 \
**Time:** 3m 12s + ~3m fixing bugs \
**Cost:** 8 credits \
**App Quality:** 9/10 \
**Code Quality:** 9.5/20

**Notes:**

- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot and before navigating to the DB selection screen
    - Only fetches DB records on DB load
    - Embeds schema into code
    - Initializes `./db/` directory on boot
    - Reuses a single `regexp` struct for new DB name validation
    - Uses `strings.Builder` for rendering
- Cons:
    - Doesn't use a constant for the `.db` extension
    - Doesn't return DB records using a custom struct; renders them directly to `string`
    - Identical dummy records for each DB
    - No unit tests
    - Non-idiomatic import blocks
    - Adding more seed data requires changing the code, since it's hardcoded to have 15 dummy `string`s
    - Compilation errors:
        - unused `path/filepath` import in `main.go`
        - unused `path/filepath` import in `db.go`
        - missing `bubbles/textinput` import in `update.go`
        - missing `lipgloss` import in `model.go`
        - `CreateDB()` expects schema as argument, which `createCmd` it tries to get from a `db.go` function that doesn't exist
        - `loadCmd()` doesn't handle DB names correctly

#### DeepSeek V4.1 Flash

**Release Date:** September 9, 2026 \
**Time:** 1m 48s \
**Cost:** 6 credits \
**App Quality:** 8/10 \
**Code Quality:** 14.5/20

**Notes:**

- Process:
    - DeepSeek V4.1 Flash included a `.gitignore` file
    - DeepSeek V4.1 Flash uses `.sqlite` extension instead of `.db`
- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot (okay because DB can't be unloaded)
    - Only fetches DB records on DB load
    - Embeds schema into code
    - Uses a constant for the `.sqlite` extension
    - Initializes `./db/` directory before reading from / writing to it
    - Reuses a single `regexp` struct for new DB name validation
    - Uses `strings.Builder` for rendering
    - Compiles without error
- Cons:
    - Embeds schema as `embed.FS` instead of `[]byte` / `string`
    - Doesn't return DB records using a custom struct; renders them directly to `string`
    - Identical dummy records for each DB
    - No unit tests
    - Code is a little more convoluted and ugly than it needs to be

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

#### Gemini 3.8 Flash

**Release Date:** September 1, 2026 \
**Time:** 25s \
**Cost:** 9 credits \
**App Quality:** 10/10 \
**Code Quality:** 15/20

**Notes:**

- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot, before navigating to the DB selection screen, and after writing to it
    - Only fetches DB records on DB load
    - Embeds schema into code
    - Initializes `./db/` directory on boot and before reading from it
    - Uses a custom `Record` struct for returning fetched DB records
    - Uses `strings.Builder` for rendering
    - Compiles without error
- Cons:
    - Doesn't use a constant for the `.db` extension
    - Doesn't reuse a single `regexp` struct during DB name validation
    - Identical dummy records for each DB
    - No unit tests
    - Non-idiomatic import blocks
    - Mixes value and pointer receivers

### Meta AI

#### Muse Spark 1.2

**Release Date:** August 4, 2026 \
**Time:** 18s \
**Cost:** 8 credits \
**App Quality:** 4/10 \
**Code Quality:** 12/20

**Notes:**

- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot (this is okay because DBs can't be unloaded)
    - Only fetches DB records on DB load
    - Initializes `./db/` directory on boot and before reading from / writing to it
    - Reuses a single `regexp` struct for new DB name validation
    - Uses a custom `Stats` struct for DB summary
    - Uses `strings.Builder` for rendering
    - Handles multiple DB file extensions
    - Compiles without error
- Cons:
    - Does not embed schema into code
    - Doesn't use a constant for the `.db` extension
    - `Stats` struct holds record values as rendered `string`
    - Identical dummy records for each DB
    - No unit tests
    - Non-idiomatic import blocks
    - Code would be a little nicer with helpers in `model.go`
    - `schema.sql` has a typo (`name TEXT NOT NOT NULL`) that prevents new DBs from being initialized

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

### Moonshot AI

#### Kimi K2.6

**Release Date:** April 20, 2026 \
**Time:** 5m 14s \
**Cost:** 22 credits \
**App Quality:** 9/10 \
**Code Quality:** 14/20

**Notes:**

- Process:
    - Kimi K2.6 had follow-up questions, documented in [RESPONSE.md](./kimi-k2.6/RESPONSE.md)
- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot and before navigating to the DB selection screen
    - Only fetches DB records on DB load
    - Embeds schema into code
    - Initializes `./db/` directory on boot and before writing to it
    - Uses custom `Record` struct for DB records
    - Uses `strings.Builder` for rendering
    - Compiles without error
- Cons:
    - Embeds schema as `embed.FS` instead of `[]byte` / `string`
    - Doesn't use a constant for the `.db` extension
    - Doesn't reuse `regexp` struct
    - Identical dummy records for each DB
    - No unit tests
    - Non-idiomatic import blocks
    - `Model.close()` and `Model.closeDB()` could easily be unified
    - Could stand to split up `model.go` into several files

#### Kimi K3

**Release Date:** July 15, 2026 \
**Time:** 49s \
**Cost:** 20 credits \
**App Quality:** 7/10 \
**Code Quality:** 16/20

**Notes:**

- Process:
    - Kimi K3 saves DBs using `.sqlite` as the extension instead of `.db`
- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot (this is okay since DBs can't be unloaded)
    - Only fetches DB records on DB load
    - Embeds schema into code
    - Initializes `./db/` directory on boot
    - Uses a simple ASCII range scan for DB name validation
    - Uses `strings.Builder` for rendering
    - Compiles without error
- Cons:
    - Doesn't use a constant for the `.sqlite` extension
    - Doesn't return DB records using a custom struct; renders them directly to `string`
    - Identical dummy records for each DB
    - No unit tests
    - Has a constant for the `./db` directory in `db.go` but doesn't use it in `main.go`
    - Could stand to split up `model.go` into several files

### OpenAI

#### GPT 5.6 Terra

**Release Date:** July 8, 2026 \
**Time:** 3m 8s \
**Cost:** 8 credits \
**App Quality:** 9/10 \
**Code Quality:** 16/20

**Notes:**

- Process:
    - GPT 5.6 Terra included a `.gitignore` file
- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot and before navigating to the DB selection screen
    - Only fetches DB records on DB load
    - Embeds schema into code
    - Uses a constant for the `.db` extension
    - Initializes `./db/` directory before writing to it
    - Reuses a single `regexp` struct for new DB name validation
    - Uses a custom `Table` struct for DB summary
    - Uses `strings.Builder` for rendering
    - Compiles without error
- Cons:
    - Embeds schema as `embed.FS` instead of `[]byte` / `string`
    - `Table` struct holds record values as rendered `string`s
    - Identical dummy records for each DB
    - No unit tests
    - `item` struct having an `isNew` field is inefficient; should be index-based
    - `Model.Update()` logic repeats logic for `tea.KeyMsg`

#### GPT 5.6 Terra Pro

**Release Date:** July 8, 2026 \
**Time:** 2m 47s \
**Cost:** 10 credits \
**App Quality:** 10/10 \
**Code Quality:** 15.5/16

**Notes:**

- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot, before navigating to the DB selection screen, and on-demand
    - Only fetches DB records on DB load
    - Embeds schema into code
    - Uses a constant for the `.db` extension
    - Initializes `./db/` directory on boot and before writing to it
    - Reuses a single `regexp` struct for new DB name validation
    - Uses a custom `Summary` struct for DB summary
    - Uses `strings.Builder` for rendering
    - Compiles without error
- Cons:
    - Embeds schema as `embed.FS` instead of `[]byte` / `string`
    - `Summary` struct holds record values as rendered `strings`
    - Identical dummy records for each DB
    - No unit tests
    - `Makefile` is over-parameterized
    - `DirName` is already public, doesn't need `Dir()` as a getter
    - Checking whether the current list of existing DBs contains a name during validation is pointless if it's just going to check whether a DB with that name exists in the directory afterwards anyways to confirm
    - Over-abundance of helpers in `app.go`:
        - `openNameScreen()`
        - `dbLoadedMsg` + `loadCmd`
        - `dbCreatedMsg` + `createCmd`

#### GPT 5.6 Sol

**Release Date:** July 8, 2026 \
**Time:** 2m 13s \
**Cost:** 22 credits \
**App Quality:** 10/10 \
**Code Quality:** 16/20

**Notes:**

- Process:
    - GPT 5.6 Sol included a `.gitignore` file
- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot and before reading from / writing to it
    - Only fetches DB records on DB load
    - Embeds schema into code
    - Uses a constant for the `.db` extension
    - Initializes `./db/` directory on boot and before reading from / writing to it
    - Reuses a single `regexp` struct for new DB name validation
    - Uses `strings.Builder` for rendering
    - Compiles without error
- Cons:
    - Embeds schema as `embed.FS` instead of `[]byte` / `string`
    - Doesn't return DB records using a custom struct; renders them directly to `string`
    - Identical dummy records for each DB
    - No unit tests
    - Could stand to split up `app.go` into several files

#### GPT 5.6 Sol Pro

**Release Date:** July 8, 2026 \
**Time:** 1m 58s \
**Cost:** 28 credits \
**App Quality:** 9/10 \
**Code Quality:** 15.5/20

**Notes:**

- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot, before navigating to the DB selection screen, and on DB load (for some reason)
    - Only fetches DB records on DB load
    - Embeds schema into code
    - Uses a constant for the `.db` extension
    - Initializes `./db/` directory before reading from / writing to it
    - Reuses a single `regexp` struct for new DB name validation
    - Uses a custom `Snapshot` struct for DB summary
    - Uses `strings.Builder` for rendering
    - Compiles without error
- Cons:
    - `Snapshot` struct holds records as rendered `strings`
    - Identical dummy records for each DB
    - No unit tests
    - Some repetition in the update helpers that could be removed by using existing Bubble Tea bubbles or rewriting `Model.Update()`
    - Could stand to split up `tui.go` into several files

#### GPT 6 Astra

**Release Date:** September 3, 2026 \
**Time:** 45s \
**Cost:** 82 credits \
**App Quality:** 9/10 \
**Code Quality:** 15/20

**Notes:**

- Process:
    - GPT 6 Astra included a `.gitignore` file
- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot and before navigating to the DB selection screen
    - Only fetches DB records on DB load
    - Embeds schema into code
    - Initializes `./db/` directory before reading from it
    - Reuses a single `regexp` struct for new DB name validation
    - Uses a custom `Snapshot` struct for DB summary
    - Uses `strings.Builder` for rendering, with extra efficient writes
    - Compiles without error
    - Atomic file creation & record fetch
- Cons:
    - Doesn't use a constant for the `.db` extension
    - `Snapshot` struct holds records as rendered `strings`
    - Identical dummy records for each DB
    - No unit tests
    - `Makefile` paramaterization could be better and not sure why it calls `gofmt` instead of `go fmt`
    - No independent name validation helper
    - It goes through a whole unnecessary rigamarole just to open a local sqlite DB

### Xiaomi

#### MiMo V2.6 Flash

**Release Date:** September 21, 2026 \
**Time:** 3m 19s + ~1m debugging \
**Cost:** 3 credits \
**App Quality:** 9/10 \
**Code Quality:** 13/20

**Notes:**

- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot, before navigating to the DB selection screen, and after creating a new DB
    - Only fetches DB records on DB load
    - Embeds schema into code
    - Uses a constant for the `.db` extension
    - Initializes `./db/` directory before writing to it
    - Reuses a single `regexp` struct for new DB name validation
    - Uses `strings.Builder` for rendering
- Cons:
    - Embeds schema as `embed.FS` instead of `[]byte` / `string`
    - Doesn't return DB records using a custom struct; renders them directly to `string`
    - Identical dummy records for each DB
    - No unit tests
    - Compile errors
        - Both struct and function ambiguously named `Preview`
        - `main.go` tries to run `ui.New()` instead of passing it to `tea.NewProgram()`

#### MiMo V2.6 Pro

**Release Date:** September 21, 2026 \
**Time:** 7m 58s + ~1m debugging \
**Cost:** 9 credits \
**App Quality:** 9/10 \
**Code Quality:** 13/20

**Notes:**

- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot and after creating a new DB
    - Only fetches DB records on DB load
    - Embeds schema into code
    - Uses a constant for the `.db` extension
    - Initializes `./db/` directory before reading from /writing to it
    - Reuses a single `regexp` struct for new DB name validation
    - Uses a custom `Table` struct for DB summary
    - Uses `strings.Builder` for rendering
    - Generic seed data generation code
- Cons:
    - Embeds schema as `embed.FS` instead of `[]byte` / `string`
    - `Table` struct holds record values as rendered `string`s
    - Identical dummy records for each DB
    - No unit tests
    - Compile errors
        - `bubbletea` package version doesn't exist causing build failure
        - Missing `lipgloss` package import in `viewer.go`
        - Unused `bubbles/textinput` package import in `create.go`

### Z AI

#### GLM 5.3

**Release Date:** July 17, 2026 \
**Time:** DNF \
**Cost:** N/A \
**App Quality:** N/A \
**Code Quality:** N/A

**Notes:**

- Multiple attempts failed to produce an answer. The model would spin for 5+ minutes before either failing or producing 2/3 files and _then_ failing.

#### GLM 5.3 Flash

**Release Date:** August 25, 2026 \
**Time:** 11s \
**Cost:** 1 credit \
**App Quality:** 8/10 \
**Code Quality:** 13/20

**Notes:**

- Pros:
    - Manages view using internal state machine
    - Only fetches the contents of `./db/` on boot (this is okay because DBs can't be unloaded)
    - Initializes `./db/` directory on boot
    - Reuses a single `regexp` struct for new DB name validation
    - Uses custom `Record` struct for DB records
    - Compiles without error
- Cons:
    - Fetches DB records on screen rerender
    - Does not embed schema into code
    - Doesn't use a constant for the `.db` extension
    - Doesn't use a `strings.Builder` for rendering
    - Identical dummy records for each DB
    - No unit tests
