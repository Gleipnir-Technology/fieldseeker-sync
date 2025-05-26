# FieldSeeker Sync Bridge

This is a program that synchronizes information between a Postgres database and FieldSeeker.
This is done to allow for adding additional items to the schema and vastly speeding up operations.

## Hacking

First, start a database:

```sh
./start-database.sh
```

Build the code:

```
nix-shell
go build
```

Then run it connecting to the database

```
env DATABASE_URL=postgresql://fieldseeker:letmein@localhost:5432 ./fieldseeker-sync```

Check on the status of migrations

```
env GOOSE_DRIVER=postgres GOOSE_DBSTRING="user=fieldseeker dbname=fieldseeker password=letmein" goose status
```
