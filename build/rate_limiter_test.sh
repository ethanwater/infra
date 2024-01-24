#!/bin/bash

while true
do
	url="http://127.0.0.1:8080/bella/2FA?action=generate"
	curl "$url"
done
