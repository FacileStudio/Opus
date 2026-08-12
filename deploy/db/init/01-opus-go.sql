-- The Go backend keeps its own schema on its own database (suite int64 ids,
-- porte's identity contract), while the legacy TS API owns the CUID2 "opus"
-- database. This runs once when the postgres data volume is first created;
-- an existing opus-db-data volume needs the database created by hand
-- (createdb -h localhost -U opus opus_go).
CREATE DATABASE opus_go;
