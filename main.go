package main

import (
  "github.com/mattburman/tesco/cmd"
  _ "github.com/mattn/go-sqlite3"
)

func main() {
  cmd.Execute()
}
