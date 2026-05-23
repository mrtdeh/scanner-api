#!/bin/bash

curl -X POST -F "file=@/path/to/file.txt" http://localhost:8080/scan


curl -X GET  http://localhost:8080/history
