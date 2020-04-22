# pump
pump random data into tables

Build:  `make install` or `make installlinux`

examples:

```bash
$ pump -h
Built on 2019-12-13 17:24:45 +0800 by go version go1.13.5 darwin/amd64 from sha1 2019-12-05 22:44:08 +0800 @293ac94d9d8f5bad8d92405a60786a0a9ce9375d @
  -b, --batch int           batch rows (default 1000)
  -d, --ds string           eg. 
                                MYSQL_PWD=8BE4 mysql -h 127.0.0.1 -P 9633 -u root
                                mysql -h 127.0.0.1 -P 9633 -u root -p8BE4
                                mysql -h 127.0.0.1 -P 9633 -u root -p8BE4 -Dtest
                                mysql -h127.0.0.1 -u root -p8BE4 -Dtest
                                127.0.0.1:9633 root/8BE4
                                127.0.0.1 root/8BE4
                                127.0.0.1:9633 root/8BE4 db=test
                                root:8BE4@tcp(127.0.0.1:9633)/?charset=utf8mb4&parseTime=true&loc=Local
                            
  -e, --eval                eval sqls execution in REPL mode
  -f, --fmt string          query sql execution result printing format(txt/markdown/html/csv) (default "txt")
  -g, --goroutines int      go routines to pump for each table (default 1)
  -h, --help                help
      --onerr string        retry on error or not (default "retry")
      --pprof-addr string   pprof address to listen on, not activate pprof if empty, eg. --pprof-addr localhost:6060
      --retry int           retry max times
  -r, --rows int            pump total rows (default 1000)
      --sleep string        sleep after each batch, eg. 10s (ns/us/µs/ms/s/m/h)
  -s, --sqls string         execute sqls, separated by ;
  -t, --tables string       pump tables, separated by ,
  -V, --verbose int         verbose details(0 off, 1 abbreviated, 2 full

./pump --ds "xx:yyy@tcp(a.b.c.d:3306)/e?charset=utf8mb4&parseTime=true&loc=Local&timeout=10s&writeTimeout=10s&readTimeout=10s" -t test_ecdocument_signatory_uuid --rows 100000
test_ecdocument_signatory_uuid pumped 334(100.00%) rows cost 689ms/5m10s901ms

./pump --ds "xx:yyy@tcp(a.b.c.d:3306)/e?charset=utf8mb4&parseTime=true&loc=Local&timeout=10s&writeTimeout=10s&readTimeout=10s" -t test_ecdocument_signatory_snow --rows 100000
test_ecdocument_signatory_snow pumped 333(100.00%) rows cost 561ms/3m55s771ms
```

or by ENV variables:

```bash
export PUMP_DS="xx:yyy@tcp(a.b.c.d:3306)/e?charset=utf8mb4&parseTime=true&loc=Local&timeout=10s&writeTimeout=10s&readTimeout=10s"
export PUMP_TABLES="sc_ecdocument,sc_ecdocument_signatory"
./pump --rows 100000
```

局域网数据库，速度快很多：

```bash
$ tail uuid.out
test_ecdocument_signatory_uuid pumped 1000(98.67%) rows cost 873ms/1m56s64ms
test_ecdocument_signatory_uuid pumped 1000(99.67%) rows cost 837ms/1m56s901ms
test_ecdocument_signatory_uuid pumped 334(100.00%) rows cost 290ms/1m57s192ms
$ tail snow.out
test_ecdocument_signatory_snow pumped 1000(99.33%) rows cost 582ms/1m11s614ms
test_ecdocument_signatory_snow pumped 333(99.67%) rows cost 153ms/1m11s768ms
test_ecdocument_signatory_snow pumped 333(100.00%) rows cost 143ms/1m11s912ms
```


pump options in columns comment refer to [faker tag](https://github.com/bingoohuang/faker/blob/master/WithStructTag.md), 
like column comments in the following, we defined pump options like `pump:"uuid_hyphenated"`, `pump:"enum=0/1/2"`, `pump:"china_mobile_number"`:

```sql
CREATE TABLE `test_ecdocument_signatory_snow`
(
    `syid`               bigint(20)   NOT NULL COMMENT 'pump:"snow"',
    `ecdocumentid`       bigint(20)   DEFAULT NULL COMMENT '合同ID pump:"snow"',
    `signnum`            int(11)      DEFAULT NULL COMMENT '签署顺序',
    `signuserid`         varchar(32)  DEFAULT NULL COMMENT '签署人id',
    `signusername`       varchar(200) DEFAULT NULL COMMENT '签署人名称',
    `certificatesnum`    varchar(100) NOT NULL COMMENT '证件号',
    `signtype`           varchar(4)   NOT NULL COMMENT '签署类型：0签署、1抄送  pump:"enum=0/1"',
    `operator`           varchar(100) DEFAULT NULL COMMENT '经办人',
    `operatorphone`      varchar(20)  DEFAULT NULL COMMENT '经办人手机 pump:"china_mobile_number"',
    `signtime`           datetime     DEFAULT NULL COMMENT '签署时间',
    `signstate`          varchar(4)   DEFAULT NULL COMMENT '签署状态：0等待签署、1已签署、2正在签署 pump:"enum=0/1/2"',
    `orgid`              varchar(32)  NOT NULL COMMENT '所属标识(企业ID/个人ID)',
    `issender`           varchar(4)   DEFAULT NULL COMMENT '是否发起方 1是 0否 pump:"enum=0/1"',
    `totalsign`          int(4)       DEFAULT NULL COMMENT '签署总数',
    `currentsign`        int(4)       DEFAULT NULL COMMENT '当前签署数',
    `dnumsignname`       varchar(200) DEFAULT NULL COMMENT '合同编号和签署方名称',
    `orgname`            varchar(100) NOT NULL COMMENT '所属方名称',
    `signatorytype`      varchar(4)   DEFAULT NULL COMMENT '所属方类型0个人1企业 pump:"enum=0/1"',
    `template_sign_code` varchar(32)  DEFAULT NULL COMMENT '模板中签署方标识',
    `processtype`        varchar(4)   DEFAULT NULL COMMENT '签署处理类别 1：顺序签署；2：不分先后，冗余字段查询时使用 pump:"enum=1/2"',
    PRIMARY KEY (`syid`),
    KEY `FHINDEX_snow` (`orgid`, `signstate`, `signtype`, `dnumsignname`(191))
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='签署方表';


CREATE TABLE `test_ecdocument_signatory_uuid`
(
    `syid`               varchar(32)  NOT NULL COMMENT 'pump:"uuid_hyphenated"',
    `ecdocumentid`       varchar(32)  DEFAULT NULL COMMENT '合同ID pump:"uuid_hyphenated"',
    `signnum`            int(11)      DEFAULT NULL COMMENT '签署顺序',
    `signuserid`         varchar(32)  DEFAULT NULL COMMENT '签署人id',
    `signusername`       varchar(200) DEFAULT NULL COMMENT '签署人名称',
    `certificatesnum`    varchar(100) NOT NULL COMMENT '证件号',
    `signtype`           varchar(4)   NOT NULL COMMENT '签署类型：0签署、1抄送  pump:"enum=0/1"',
    `operator`           varchar(100) DEFAULT NULL COMMENT '经办人',
    `operatorphone`      varchar(20)  DEFAULT NULL COMMENT '经办人手机 pump:"china_mobile_number"',
    `signtime`           datetime     DEFAULT NULL COMMENT '签署时间',
    `signstate`          varchar(4)   DEFAULT NULL COMMENT '签署状态：0等待签署、1已签署、2正在签署 pump:"enum=0/1/2"',
    `orgid`              varchar(32)  NOT NULL COMMENT '所属标识(企业ID/个人ID)',
    `issender`           varchar(4)   DEFAULT NULL COMMENT '是否发起方 1是 0否 pump:"enum=0/1"',
    `totalsign`          int(4)       DEFAULT NULL COMMENT '签署总数',
    `currentsign`        int(4)       DEFAULT NULL COMMENT '当前签署数',
    `dnumsignname`       varchar(200) DEFAULT NULL COMMENT '合同编号和签署方名称',
    `orgname`            varchar(100) NOT NULL COMMENT '所属方名称',
    `signatorytype`      varchar(4)   DEFAULT NULL COMMENT '所属方类型0个人1企业 pump:"enum=0/1"',
    `template_sign_code` varchar(32)  DEFAULT NULL COMMENT '模板中签署方标识',
    `processtype`        varchar(4)   DEFAULT NULL COMMENT '签署处理类别 1：顺序签署；2：不分先后，冗余字段查询时使用 pump:"enum=1/2"',
    PRIMARY KEY (`syid`),
    KEY `FHINDEX_uuid` (`orgid`, `signstate`, `signtype`, `dnumsignname`(191))
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='签署方表';
```

## Eval mode

```bash
$ pump -d "127.0.0.1 user/pass db=mydb" -V -e
INFO[0000] dataSourceName:user:pass@tcp(127.0.0.1:3306)/mydb?charset=utf8mb4&parseTime=true&loc=Local
Enter your sql (empty to re-execute) : select * from t_server
2020/02/13 10:55:16 SQL: select * from t_server
2020/02/13 10:55:16 cost: 610.248µs
+---+----+-------------+------+-------+
| # | ID | SERVER      | PORT | STATE |
+---+----+-------------+------+-------+
| 1 | 1  | 192.168.1.1 | 8001 | 1     |
| 2 | 2  | 192.168.1.2 | 8001 | 1     |
+---+----+-------------+------+-------+
Enter your sql (empty to re-execute) :
2020/02/13 10:55:25 SQL: select * from t_server
2020/02/13 10:55:25 cost: 6.6981ms
+---+----+-------------+------+-------+
| # | ID | SERVER      | PORT | STATE |
+---+----+-------------+------+-------+
| 1 | 1  | 192.168.1.1 | 8001 | 1     |
| 2 | 2  | 192.168.1.2 | 8001 | 1     |
+---+----+-------------+------+-------+
Enter your sql (empty to re-execute) :
```

## Thanks

1. [MySQL random data loader](https://github.com/Percona-Lab/mysql_random_data_load)
1. [Go访问MySQL异常排查及浅析其超时机制](https://cloud.tencent.com/developer/article/1390124)
