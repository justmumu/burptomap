#!/bin/bash

REQUESTS_FOLDER="{{ .REQS_FOLDER }}"
OUTPUTS_DIR="{{ .OUTPUTS_FOLDER }}"

for request_file in "$REQUESTS_FOLDER"/*
do
    reqNumber=$(echo "$request_file" | awk -F"/" '{print $NF}' | cut -d"-" -f2 | cut -d"." -f1)
    
    sqlmap -r ${request_file} {{.SQLMAP_COMMANDS}} | tee ${OUTPUTS_DIR}/result-${reqNumber}.output
done