package message_broker

type ProducerSpec struct {
	SendRetries        int `json:"send_retries"`
	ConnMaxLifetimeSec int `json:"conn_max_lifetime_sec"`
	ConnMaxIdleTimeSec int `json:"conn_max_idle_time_sec"`
}

type ConsumerSpec struct {
	GroupId            string `json:"group_id"`
	ReadRetries        int    `json:"read_retries"`
	WorkersAmount      int    `json:"workers_amount"`
	MaxIdleCons        int    `json:"max_idle_cons"`
	MaxOpenCons        int    `json:"max_open_cons"`
	ConnMaxLifetimeSec int    `json:"conn_max_lifetime_sec"`
	ConnMaxIdleTimeSec int    `json:"conn_max_idle_time_sec"`
}

type Settings struct {
	Brokers  []string     `json:"brokers"`
	Producer ProducerSpec `json:"producer"`
	Consumer ConsumerSpec `json:"consumer"`
}
