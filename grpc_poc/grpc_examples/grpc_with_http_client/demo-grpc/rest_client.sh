
unset http_proxy
curl -H "login:john" -H "password:doe" -X POST -d '{"greeting":"foo"}' 'http://localhost:7778/1/ping'

