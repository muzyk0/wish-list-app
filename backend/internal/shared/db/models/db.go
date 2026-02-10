package db

import "wish-list/internal/app/database"

// Re-export types from app/database for backward compatibility during migration
type DB = database.DB
type Executor = database.Executor

var New = database.New
var UUIDToString = database.UUIDToString
var StringToUUID = database.StringToUUID
var TextToString = database.TextToString
var StringToText = database.StringToText
var BoolToBool = database.BoolToBool
var BoolToPgBool = database.BoolToPgBool
var NumericToFloat64 = database.NumericToFloat64
