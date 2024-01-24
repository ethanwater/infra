#!/bin/bash

counter=0

while [ "$counter" -lt 15 ]
do
	url="http://127.0.0.1:8080/bella/2FA?action=generate"
	curl "$url"

	((counter++))
done
