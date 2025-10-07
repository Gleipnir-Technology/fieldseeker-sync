# FieldSeeker Sync Bridge

This is a program that synchronizes information between a Postgres database and FieldSeeker.
This is done to allow for adding additional items to the schema and vastly speeding up operations.

## Updating the schema

If you get a message from the `export` process like:

```
Sep 20 03:10:37 sync.nidus.cloud full-export[1082202]: 2025/09/20 03:10:37 Need to get 154018 records for layer 'MosquitoInspection'
Sep 20 03:11:51 sync.nidus.cloud full-export[1082202]: 2025/09/20 03:11:51 Type: <nil>        key: SRID        value: <nil>        row: c6a26cef-29e5-499b-9422-251e63fc185e
Sep 20 03:11:51 sync.nidus.cloud full-export[1082202]: 2025/09/20 03:11:51 Need type update.
```

Then that means you need to update the schema. In my example here, I need to update `MosquitoInspection`.

We start by downloading the schema. You do that through the `download-schema` tool

```sh
./result/bin/download-schema
```

This will populate a number of files in `schema`. We care about `schema/MosquitoInspection.json`. If you just look at the file it's pretty hard to read because it has no whitespace formatting. You can use `jq` or `python3 -m json.tool schema/MosquitoInspection.json`.

Eventually we're going to run `tools/generate-db-schema.py`. By default this will read all the files in `schema` and generate a single full schema file in `database/migrations/current-fieldseeker-schema`. This is useful for generating diffs.

```sh
python3 tools/generate-db-schema.py
```

Then `git diff` will show what's changed. The tough bit is now you have to create a schema migration file out of the changes. In my case I have:

```
diff --git a/database/migrations/current-fieldseeker-schema b/database/migrations/current-fieldseeker-schema
index c9df235..307622c 100644
--- a/database/migrations/current-fieldseeker-schema
+++ b/database/migrations/current-fieldseeker-schema
@@ -196,6 +196,7 @@ CREATE TABLE FS_MosquitoInspection (
        POLYGONLOCID TEXT,
        POSDIPS INT2,
        POSITIVECONTAINERCOUNT INT2,
+       PTAID TEXT,
        PUPAEPRESENT INT2,
        RAINGAUGE DOUBLE PRECISION,
        RECORDSTATUS INT2,
@@ -482,6 +483,7 @@ CREATE TABLE FS_RodentLocation (
        Editor TEXT,
        GlobalID TEXT,
        HABITAT TEXT,
+       JURISDICTION TEXT,
        LASTINSPECTACTION TEXT,
        LASTINSPECTCONDITIONS TEXT,
        LASTINSPECTDATE BIGINT,
@@ -745,6 +747,7 @@ CREATE TABLE FS_TimeCard (
        OBJECTID INTEGER PRIMARY KEY,
        POINTLOCID TEXT,
        POLYGONLOCID TEXT,
+       RODENTLOCID TEXT,
        SAMPLELOCID TEXT,
        SRID TEXT,
        STARTDATETIME BIGINT,
```

This tells me everything I need to know. I need to add 3 columns, `FS_MosquitoInspection.PTAID`, `FS_RodentLocation.JURISDICTION`, and `FS_TimeCard.RODENTLOCID`.

We'll create a new migration for that. At this point, that's going to be `database/migrations/00020_ptaid_jurisdiction_rodentlocid`. You can look at the source to see how to map the changes to SQL statements.

Make sure to remember the history tables.

You can use `goose` to check the current status and go up and down in migrations to make sure you get things right:
```sh
cd database/migrations
GOOSE_DRIVER=postgres GOOSE_DBSTRING="dbname=fieldseeker-sync host=/var/run/postgresql" goose status
GOOSE_DRIVER=postgres GOOSE_DBSTRING="dbname=fieldseeker-sync host=/var/run/postgresql" goose up
GOOSE_DRIVER=postgres GOOSE_DBSTRING="dbname=fieldseeker-sync host=/var/run/postgresql" goose down
```

## Hacking

First, start a database:

```sh
./start-database.sh
```

Build the code:

```
nix develop
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
