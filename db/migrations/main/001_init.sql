CREATE TABLE `user_shard` (
  `user_id`    int(11) unsigned NOT NULL,
  `shard_id`   tinyint(3) unsigned NOT NULL,
  
  PRIMARY KEY `user_id` (`user_id`)
) CHARACTER SET utf8 COLLATE utf8_general_ci ENGINE=InnoDB;