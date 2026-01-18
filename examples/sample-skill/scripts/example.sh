#!/bin/bash

# Example Script for Sample Skill
# This script demonstrates how to include executable scripts in a skill

echo "Hello from sample-skill!"
echo "This is an example script that can be referenced by the AI."

# You can include any shell commands or logic here
# The AI can read and understand this script to help with tasks

# Example: Check if a file exists
if [ -f "$1" ]; then
    echo "File $1 exists"
    wc -l "$1"
else
    echo "Usage: $0 <filename>"
    exit 1
fi
