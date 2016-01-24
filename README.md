## GORM generator

## Basic idea

If you have an existing database, manually typing out structs for all of your tables kinda sucks.
These structs are mostly a lot of boilerplate code, so generating them by reading the table information
makes sense to me.

## Limitations

#### Postgres only

I've only tested this with postgres.

#### Super limited column types

The column types are pretty limited, just based on what I've encountered in my own tables.

#### No customization

No custom PK columns, etc. right now
