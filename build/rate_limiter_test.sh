counter=0

while [ "$counter" -lt 15 ]
do
    url="http://127.0.0.1:8080/bella/2FA?action=generate"
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url")

    if [ "$response" -ne 200 ]; then
        echo "Request failed with status code $response. Exiting script."
        break
    fi

    ((counter++))
done
