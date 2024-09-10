CREATE TABLE `product_photo` (
  `id`         tinyint(3) unsigned NOT NULL,
  `advert_id`  int(11) unsigned NOT NULL, 
  `url`        varchar(255) NOT NULL,
  `url_small`  varchar(255) NOT NULL DEFAULT '',
  `url_medium` varchar(255) NOT NULL DEFAULT '',
  `url_big`    varchar(255) NOT NULL DEFAULT '',
  `position`   tinyint(3) unsigned NOT NULL,
  
  PRIMARY KEY `advert_id,id`  (`advert_id`,`id`)
) CHARACTER SET utf8 COLLATE utf8_general_ci ENGINE=InnoDB;