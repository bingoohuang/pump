# pump
pump random data into tables


```bash
$ export PUMP_DATA_SOURCE="user:pass@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=true&loc=Local"
$ export PUMP_TABLE="t_person"
$ export PUMP_ROWS=10000
$ ./pump
 begin to pump 10000 rows to table sc_ecdocument_signatory
 10000 rows added, cost 12.964908615s
 complete! total 10000 rows added
```


pump options in columns comment refer to [faker tag](https://github.com/bingoohuang/faker/blob/master/WithStructTag.md), 
like column comments in the following, we defined pump options like `pump:"uuid_hyphenated"`, `pump:"enum=0/1/2"`, `pump:"china_mobile_number"`:

```sql
CREATE TABLE `sc_ecdocument_signatory` (
  `syid` varchar(32) NOT NULL COMMENT 'pump:"uuid_hyphenated"',
  `ecdocumentid` varchar(32) DEFAULT NULL COMMENT '合同ID',
  `signnum` int(11) DEFAULT NULL COMMENT '签署顺序',
  `signuserid` varchar(32) DEFAULT NULL COMMENT '签署人id',
  `signusername` varchar(200) DEFAULT NULL COMMENT '签署人名称',
  `certificatesnum` varchar(100) NOT NULL COMMENT '证件号',
  `signtype` varchar(4) NOT NULL COMMENT '签署类型：0签署、1抄送  pump:"enum=0/1"',
  `operator` varchar(100) DEFAULT NULL COMMENT '经办人',
  `operatorphone` varchar(20) DEFAULT NULL COMMENT '经办人手机 pump:"china_mobile_number"',
  `signtime` datetime DEFAULT NULL COMMENT '签署时间',
  `signstate` varchar(4) DEFAULT NULL COMMENT '签署状态：0等待签署、1已签署、2正在签署 pump:"enum=0/1/2"',
  `orgid` varchar(32) NOT NULL COMMENT '所属标识(企业ID/个人ID)',
  `issender` varchar(4) DEFAULT NULL COMMENT '是否发起方 1是 0否 pump:"enum=0/1"',
  `totalsign` int(4) DEFAULT NULL COMMENT '签署总数',
  `currentsign` int(4) DEFAULT NULL COMMENT '当前签署数',
  `dnumsignname` varchar(200) DEFAULT NULL COMMENT '合同编号和签署方名称',
  `orgname` varchar(100) NOT NULL COMMENT '所属方名称',
  `signatorytype` varchar(4) DEFAULT NULL COMMENT '所属方类型0个人1企业 pump:"enum=0/1"',
  `template_sign_code` varchar(32) DEFAULT NULL COMMENT '模板中签署方标识',
  `processtype` varchar(4) DEFAULT NULL COMMENT '签署处理类别 1：顺序签署；2：不分先后，冗余字段查询时使用 pump:"enum=1/2"',
  PRIMARY KEY (`syid`),
  KEY `fk_sc_ecdocument_signatory_01` (`ecdocumentid`),
  KEY `FHINDEX` (`orgid`,`signstate`,`signtype`,`dnumsignname`(191))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='签署方表'
```


## Thanks

1. [MySQL random data loader](https://github.com/Percona-Lab/mysql_random_data_load)