import requests

re = requests.post("http://127.0.0.1:1023/logEvent", "Approve invoice num 2050")
print(re.json)