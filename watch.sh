#!/bin/bash

while inotifywait -r -e modify,create,delete .; do
  go run main.go
done
