#!/bin/bash
cd /mnt/c/laragon/www/PresensiGo/backend
go build ./cmd/api
if [ $? -eq 0 ]; then
    echo "✓ Code compiles successfully"
    exit 0
else
    echo "✗ Compilation failed"
    exit 1
fi
