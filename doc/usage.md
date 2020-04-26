# MySQL造数据工具pump使用手册

1. 下载 `https://github.com/bingoohuang/pump/releases`
1. 解压并且改名为pump，放置到$PATH环境变量定义的目录中
1. 运行 `pump -h` 查看帮助
1. 示例：
    - 向aa库的tr_f_db表插入10条数据 `pump -d "root:root@127.0.0.1:3306/aa" -t tr_f_db -r 10`
    - 向aa库的tr_f_db表插入1000条数据，每批100条 `pump -d "root:root@127.0.0.1:3306/aa" -t tr_f_db -r 1000 -b 100`
    - 直接执行SQL语句 `pump -d "root:root@127.0.0.1:3306/aa" -s "select count(*) from aa.tr_f_db"`
