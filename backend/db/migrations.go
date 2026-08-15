package db

import (
    "fmt"
    "os"
)

func ApplyMigrations() error {
    paths:=[]string{"database/migrations/002_ai_reporting.sql","backend/database/migrations/002_ai_reporting.sql"}
    var data []byte; var err error
    for _,p:=range paths{data,err=os.ReadFile(p);if err==nil{break}}
    if err!=nil{return fmt.Errorf("could not read reporting/resume migration: %w",err)}
    if _,err=DB.Exec(string(data));err!=nil{return fmt.Errorf("could not apply reporting/resume migration: %w",err)}
    return nil
}
