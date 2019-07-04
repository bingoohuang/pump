# pump
pump random data into tables


Directly by command arguments:

```bash
$ pump git:(master) ./pump -h
Built on 2019-07-02 12:40:30 +0800 by go version go1.12.6 darwin/amd64 from sha1 2019-07-01 22:30:39 +0800 @b10916c8696bba63d9de402f4a9e2f5b3da6d3af @
  -b, --batch int           batch rows (default 1000)
  -d, --datasource string   help (default "user:pass@tcp(localhost:3306)/db?charset=utf8mb4&parseTime=true&loc=Local&timeout=10s&writeTimeout=10s&readTimeout=10s")
  -g, --goroutines int      go routines to pump for each table (default 3)
  -h, --help                help
      --pprof-addr string   pprof address to listen on, not activate pprof if empty, eg. --pprof-addr localhost:6060
  -r, --rows int            pump rows (default 1000)
  -t, --tables string       pump tables

./pump -d "xx:yyy@tcp(a.b.c.d:3306)/e?charset=utf8mb4&parseTime=true&loc=Local&timeout=10s&writeTimeout=10s&readTimeout=10s" -t test_ecdocument_signatory_uuid -r 100000
test_ecdocument_signatory_uuid pumped 334(100.00%) rows cost 689ms/5m10s901ms

./pump -d "xx:yyy@tcp(a.b.c.d:3306)/e?charset=utf8mb4&parseTime=true&loc=Local&timeout=10s&writeTimeout=10s&readTimeout=10s" -t test_ecdocument_signatory_snow -r 100000
test_ecdocument_signatory_snow pumped 333(100.00%) rows cost 561ms/3m55s771ms
```

or by ENV variables:

```bash
export PUMP_DATASOURCE="xx:yyy@tcp(a.b.c.d:3306)/e?charset=utf8mb4&parseTime=true&loc=Local&timeout=10s&writeTimeout=10s&readTimeout=10s"
export PUMP_TABLES="sc_ecdocument,sc_ecdocument_signatory"
./pump -r 100000
```

局域网数据库，速度快很多：

```bash
[betaoper@beta-hetong pumptool]$ tail uuid.out
test_ecdocument_signatory_uuid pumped 1000(98.67%) rows cost 873ms/1m56s64ms
test_ecdocument_signatory_uuid pumped 1000(99.67%) rows cost 837ms/1m56s901ms
test_ecdocument_signatory_uuid pumped 334(100.00%) rows cost 290ms/1m57s192ms
[betaoper@beta-hetong pumptool]$ tail snow.out
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


## Thanks

1. [MySQL random data loader](https://github.com/Percona-Lab/mysql_random_data_load)