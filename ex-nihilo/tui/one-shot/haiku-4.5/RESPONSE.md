## Haiku 4.5

A few clarifying questions:

1. For the DB selection screen: should the user be able to go back/cancel after selecting a DB, or is selection final?
2. For DB name validation: should I reject names that match existing DBs case-insensitively, or only exact matches?
3. For the loaded DB screen: should there be a way to go back and load a different DB, or is the session tied to one DB?
4. Should the project have any configuration/initialization (e.g., creating `./db/` if it doesn't exist)?

## Me

1. The selection should be final
2. Only exact matches will suffice
3. Either is fine, the latter is preferable
4. Yes, if the `./db/` directory doesn't exist then it should be created
