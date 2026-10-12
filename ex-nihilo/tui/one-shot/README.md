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
