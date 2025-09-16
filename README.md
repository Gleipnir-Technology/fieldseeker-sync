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

Generate models for bob:

```
$ cd database
$ PSQL_DSN="postgresql://fieldseeker-sync:@?host=/var/run/postgresql&sslmode=disable" go run github.com/stephenafamo/bob/gen/bobgen-psql@latest
```

This will generate a bunch of files in `database/`.

### Hot reloading

Hot reloading is done via [air](https://github.com/air-verse/air):

```shell
$ air
```

you'll need to make sure your session has the necessary environment variables to connect to the database
