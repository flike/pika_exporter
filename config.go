package main

import (
	"io/ioutil"
	"strings"

	"github.com/pourer/pika_exporter/exporter"

	"github.com/go-ini/ini"
)

var defaultMetricConfig = `[build_info]
labels = addr,alias,os,arch_bits,pika_version,pika_git_sha,pika_build_compile_date

[server_info]
labels = addr,alias,process_id,tcp_port,config_file,server_id,role

[uptime_in_seconds]
labels = addr,alias
value = uptime_in_seconds

[thread_num]
labels = addr,alias
value = thread_num

[sync_thread_num]
labels = addr,alias
value = sync_thread_num

[db_size]
labels = addr,alias,compression
value = db_size

[used_memory]
labels = addr,alias
value = used_memory

[db_memtable_usage]
labels = addr,alias
value = db_memtable_usage

[db_tablereader_usage]
labels = addr,alias
value = db_tablereader_usage

[binlog]
labels = addr,alias,safety_purge,expire_logs_days,expire_logs_nums
value = log_size

# 由info获取到的binlog_offset为(文件编号 偏移量): 0 388
# pika_exporter对其进行了分离:
#	"文件编号"的key为: binlog_offset_filenum
#	"偏移量"的key为: binlog_offset_value
[binlog_offset]
labels = addr,alias,binlog_offset_filenum
value = binlog_offset_value

[connected_clients]
labels = addr,alias
value = connected_clients

[total_connections_received]
labels = addr,alias
value = total_connections_received

[instantaneous_ops_per_sec]
labels = addr,alias
value = instantaneous_ops_per_sec

[total_commands_processed]
labels = addr,alias
value = total_commands_processed

# 由info获取到的is_bgsaving为(是否在备份,最近一次备份开始时间,已经备份了多久): No, , 0
# pika_exporter对其进行了分离:
#	"是否在备份"的key为: is_bgsaving
#	"最近一次备份开始时间": bgsave_start_time
[is_bgsaving]
labels = addr,alias,bgsave_start_time
value = is_bgsaving

# 由info获取到的is_slots_reloading为(是否在加载,最近一次加载开始时间,已经加载了多久): No, , 0
# pika_exporter对其进行了分离:
#	"是否在加载"的key为: is_slots_reloading
#	"最近一次加载开始时间": slots_reload_start_time
[is_slots_reloading]
labels = addr,alias,slots_reload_start_time
value = is_slots_reloading

# 由info获取到的is_slots_cleaning为(是否在清理,最近一次清理开始时间,已经清理了多久): No, , 0
# pika_exporter对其进行了分离:
#	"是否在清理"的key为: is_slots_cleaning
#	"最近一次清理开始时间": slots_clean_start_time
[is_slots_cleaning]
labels = addr,alias,slots_clean_start_time
value = is_slots_cleaning

# 由info获取到的keyspace的上一次统计时间为: # Time:1970-01-01 00:00:00
# pika_exporter把其处理为Unix整型时间戳: keyspace_time
[is_scaning_keyspace]
labels = addr,alias,keyspace_time
value = is_scaning_keyspace

[is_compact]
labels = addr,alias,compact_cron,compact_interval
value = is_compact

[used_cpu_sys]
labels = addr,alias
value = used_cpu_sys

[used_cpu_user]
labels = addr,alias
value = used_cpu_user

[used_cpu_sys_children]
labels = addr,alias
value = used_cpu_sys_children

[used_cpu_user_children]
labels = addr,alias
value = used_cpu_user_children

######################################################################################################

[master_info]
labels = addr,alias
value = connected_slaves

# 由info获取到的master的slave的信息为: slave0:ip=192.168.1.1,port=57765,state=online,sid=2,lag=0
# pika_exporter把其处理为: slave_ip slave_port slave_state slave_sid slave_lag
[master_slave_info]
labels = addr,alias,slave_sid,slave_ip,slave_port,slave_state
value = slave_lag

[slave_info]
labels = addr,alias,master_host,master_port,slave_priority,slave_read_only,repl_state
value = master_link_status

[double_master_info]
labels = addr,alias,the_peer_master_server_id,the_peer_master_host,the_peer_master_port,repl_state

# 由info获取到的double_master_recv_info为(文件编号 偏移量): 0 0
# pika_exporter对其进行了分离:
#	"文件编号"的key为: double_master_recv_info_binlog_filenum
#	"偏移量"的key为: double_master_recv_info_binlog_offset
[double_master_recv_info]
labels = addr,alias,the_peer_master_host,the_peer_master_port,the_peer_master_server_id,double_master_recv_info_binlog_filenum
value = double_master_recv_info_binlog_offset

######################################################################################################

# 由info获取到的keyspace的kv数量为: kv keys:0
# pika_exporter把其处理为: kv_keys:0
[kv_keys]
labels = addr,alias
value = kv_keys

# 由info获取到的keyspace的hash-kv数量为: hash keys:0
# pika_exporter把其处理为: hash_keys:0
[hash_keys]
labels = addr,alias
value = hash_keys

# 由info获取到的keyspace的list-kv数量为: list keys:0
# pika_exporter把其处理为: list_keys:0
[list_keys]
labels = addr,alias
value = list_keys

# 由info获取到的keyspace的set-kv数量为: set keys:0
# pika_exporter把其处理为: set_keys:0
[set_keys]
labels = addr,alias
value = set_keys

# 由info获取到的keyspace的zset-kv数量为: zsetkeys:0
# pika_exporter把其处理为: zset_keys:0
[zset_keys]
labels = addr,alias
value = zset_keys`

func loadMetricConfig(fileName string) (exporter.Metrics, error) {
	var data []byte
	if fileName != "" {
		fileData, err := ioutil.ReadFile(fileName)
		if err != nil {
			return nil, err
		}
		data = fileData
	} else {
		data = []byte(defaultMetricConfig)
	}

	cfg, err := ini.Load(data)
	if err != nil {
		return nil, err
	}

	metrics := make(exporter.Metrics)
	for _, section := range cfg.Sections() {
		if section.Name() == ini.DEFAULT_SECTION {
			continue
		}

		metric := &exporter.Metric{Name: section.Name()}
		for _, key := range section.Keys() {
			keyName := strings.ToLower(key.Name())
			keyValue := strings.ToLower(key.Value())
			switch keyName {
			case "labels":
				metric.Labels = strings.Split(keyValue, ",")
			case "value":
				metric.ValueName = keyValue
			}
		}
		metrics[metric.Name] = metric
	}

	return metrics, nil
}
