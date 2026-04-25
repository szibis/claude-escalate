#!/bin/bash
read -r PROMPT
curl -s -X POST http://localhost:9000/api/hook -d "{\"prompt\":\"$PROMPT\"}"
