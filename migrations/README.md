# Migrations

All tables prefixed with `FS_` are FieldSeeker tables. We get their contents via the ArcGIS API. The only process that changes the content of these tables is the background sync process.

Each table includes a history table, `History_`, which contains the same schema as the `FS_` tables, but includes a composite primary key that includes a version number. It also includes a "created" column. The highest version is the most recent. Whenever a new row is created in the corresponding `FS_` database there is a row created at version 1 in the `History_` database. Any time data is updated in the `FS_` database a new row is added with a new version.

The history table is managed by the export process itself.

Each of the `FS_` tables is augmented with an "updated" column. This marks the last time a given row was modified, which is useful for clients to query what they need to get in order to have the latest status.
