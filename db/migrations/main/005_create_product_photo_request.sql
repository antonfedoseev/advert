CREATE TABLE `product_photo_request` (
  `id`         int(11) unsigned NOT NULL AUTO_INCREMENT,
  `photo_id`   int(11) unsigned NOT NULL, 
  `status`     tinyint(2) unsigned NOT NULL, 
  `path`       varchar(255) NOT NULL,
  `сtime`      int(11) unsigned NOT NULL DEFAULT '0',
  
  PRIMARY KEY `id`        (`id`),
  KEY         `photo_id`  (`photo_id`)
) ENGINE=InnoDB;