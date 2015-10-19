# bloomfilter
A Bloom Filter Write By Golang.

1. 依赖: 
	go get https://github.com/willf/bloom
	go get https://github.com/willf/bitset
2. 编译: 
	go build bloomf.go
3. Usage:
	./bloomf append file1 #离线build数据为二进制db
	./bloomf search key1  #离线查找指定key
	./bloomf web 9090     #启动web服务，指定某端口
4. 热增量方式：
 把增量文件mv到bloomf同目录下，文件名叫 append.txt ,系统会自动(5秒探测间隔）把增加文件加载到内存并写入db库中,增量文件会自动备份到 datas目录下，以当前时间戳为后缀
5. 容量及其他:
 目前按照4亿的集合配置的，二进制文件大小为915M,2亿文件的build时间约4分钟，系统初次加载全量db约20秒


